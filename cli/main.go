package main

import (
	"coinbit-test/core/web"
	"coinbit-test/lib/handler"
	"coinbit-test/lib/proto"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/go-chi/chi/v5"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	"github.com/lovoo/goka/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	brokers             = []string{"localhost:9092"}
	topic   goka.Stream = "example-stream"
	group   goka.Group  = "example-group"
)

func main() {
	// show file name in log
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// init goka emitter
	emitter, err := goka.NewEmitter(brokers, topic, new(codec.Bytes))
	if err != nil {
		log.Fatalf("Error creating emitter: %v", err)
	}

	// init db
	db, err := leveldb.OpenFile("database", nil)
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}
	defer db.Close()

	// init goka storage
	storage, err := storage.New(db)
	if err != nil {
		log.Fatalf("Error creating storage: %v", err)
	}

	// init handler
	handler := &handler.Handler{
		Emitter: emitter,
		Storage: storage,
	}

	// start asynchronous web server
	go startWebServer(handler)

	// run processor
	runProcessor(handler)
}

func startWebServer(handler *handler.Handler) {
	r := chi.NewRouter()
	AddRoute(r, handler, web.API)
	http.ListenAndServe(":3000", r)
}

func AddRoute(r *chi.Mux, handler *handler.Handler, route func(handler *handler.Handler, e chi.Router)) *chi.Mux {
	r.Route("/", func(c chi.Router) {
		route(handler, c)
	})
	return r
}

func runProcessor(handler *handler.Handler) {
	config := goka.DefaultConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	goka.ReplaceGlobalConfig(config)

	tmc := goka.NewTopicManagerConfig()
	tmc.Table.Replication = 1
	tmc.Stream.Replication = 1

	tm, err := goka.NewTopicManager(brokers, goka.DefaultConfig(), tmc)
	if err != nil {
		log.Fatalf("Error creating topic manager: %v", err)
	}
	defer tm.Close()
	err = tm.EnsureStreamExists(string(topic), 8)
	if err != nil {
		log.Printf("Error creating kafka topic %s: %v", topic, err)
	}

	cb := func(ctx goka.Context, msg interface{}) {
		key := ctx.Key()
		deposits := make([]proto.Deposit, 0, 10)

		b, err := handler.Storage.Get(key)
		if err != nil {
			log.Printf("Error get key %s: %v", key, err)
		}

		if b != nil {
			err = json.Unmarshal(b, &deposits)
			if err != nil {
				log.Printf("Error unmarshal key %s: %v", key, err)
			}
		}

		message, ok := msg.([]byte)
		if ok {
			var deposit proto.Deposit
			err = json.Unmarshal(message, &deposit)
			if err != nil {
				log.Printf("Error unmarshal key %s: %v", key, err)
			}

			deposit.Timestamp = timestamppb.Now()
			deposits = append(deposits, deposit)

			b, err = json.Marshal(deposits)
			if err != nil {
				log.Printf("Error marshal key %s: %v", key, err)
			}

			handler.Storage.Set(key, b)
			if err != nil {
				log.Printf("Error set key %s: %v", key, err)
			}
		}
	}

	g := goka.DefineGroup(group,
		goka.Input(topic, new(codec.Bytes), cb),
		goka.Persist(new(codec.Bytes)),
	)

	p, err := goka.NewProcessor(brokers, g)
	if err != nil {
		log.Fatalf("Error creating processor: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)
	go func() {
		defer close(done)
		if err = p.Run(ctx); err != nil {
			log.Fatalf("Error running processor: %v", err)
		} else {
			log.Printf("Processor shutdown cleanly")
		}
	}()

	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)
	<-wait
	cancel()
	<-done
}
