# Twinder Client

## Getting Started

### Build
```
make
```

### Run
```
./httpclient
```

### Deploy

Deployed on Railway.app

## Results

| Goroutine count | Throughtput (req/sec) | P99 response time (ms) |
| --------------- | --------------------- | ---------------------- |
| 1               | 593.48                | 2.00                   |
| 10              | 3525.35               | 5.00                   |
| 25              | 4284.47               | 10.00                  |
| 50              | 5324.35               | 15.00                  |
| 75              | 5428.06               | 20.00                  |
| 100             | 5365.31               | 26.00                  |

It seems like 50 goroutines is the optimal setup before significant diminishing returns.  
1.7 vCPU Peak  
20 MB Peak

## Design

![Client](results/a1-client-diagram-v2.png)

## Screenshots

### 1 goroutine
![1](results/1worker500krequests.png)

### 25 goroutines
![25](results/25workers500krequests.png)

### 50 goroutines
![50](results/50workers500krequests.png)

### 75 goroutines
![75](results/75workers500krequests.png)

### 100 goroutines
![100](results/100workers500krequests.png)

### Server throughput on 50 goroutines
![throughput](results/50workers-throughput.png)

Measured every 5 seconds. So divide by 5 to get req/sec.