# README

## Steps to try

1. Clone repository
```
git clone git@github.com:mghifariyusuf/coinbit-test.git
```

2. Download dependencies
```
make get-vendor
```

3. Generate protobuf
```
make gen-proto
```

4. Start Kafka
```
make compose
```

5. Run web server at other terminal
```
make run
```

6. Import postman collection
```
Coinbit.postman_collection.json
```