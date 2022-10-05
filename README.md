# go-fsring
Simple ring buffer

[![Go Reference](https://pkg.go.dev/badge/github.com/takanoriyanagitani/go-fsring.svg)](https://pkg.go.dev/github.com/takanoriyanagitani/go-fsring)
[![Go Report Card](https://goreportcard.com/badge/github.com/takanoriyanagitani/go-fsring)](https://goreportcard.com/report/github.com/takanoriyanagitani/go-fsring)
[![codecov](https://codecov.io/gh/takanoriyanagitani/go-fsring/branch/main/graph/badge.svg?token=Q4FFFYLX37)](https://codecov.io/gh/takanoriyanagitani/go-fsring)

###### Sample Benchmark

| bytes / document | # documents | WA |
|--:|:--:|--:|
|    16,384 | 1024 | 3400 % |
|    40,960 | 1024 | 1200 % |
|    81,920 | 1024 |  820 % |
|   163,840 | 1024 |  480 % |
|   327,680 | 1024 |  280 % |
|   512,000 | 1024 |  230 % |
| 1,024,000 | 1024 |  160 % |
| 2,097,152 | 1024 |  130 % |
