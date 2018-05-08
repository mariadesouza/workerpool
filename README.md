# Worker Pools

## Installation

```
  github.com/mariadesouza/workerpool
```

## Exported Methods

### New
  Params:
     numOfWorkers int
  Returns:
  (\*WorkerPool)

### (\*WorkerPool) AddWorkToPool()
  Params:
    func()

### (\*WorkerPool) Close()

## Example:
```
wp := workerpool.New(m)
defer wp.Close()
for i := 0; i < n; i++ {
    wp.AddWorkToPool(
			func() {
        functiontoberunconcurrently()
      }
   )
}
```
## Synopsis

## Reference
https://medium.com/@matryer/stopping-goroutines-golang-1bf28799c1cb
http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/

## Contributor
Maria De Souza
