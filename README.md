# Snake Game with AI

## Usage

```
./snakeai [-chiqmnprst] [<prefix>brain-<score>.json]
  -c    create empty brain
  -h    start in human mode
  -i int
        max number of instances in epoch (default 1000)
  -m int
        min score in epoch
  -n float
        interval of the mutation changes on the synapse weight (default 0.5)
  -p string
        prefix filename with brain
  -q    start in silent mode
  -r float
        mutation rate on the weights of synapses (default 0.1)
  -s int
        snake speed limit (default 100)
  -t int
        max snake stept without eat (default 200)
```

```
$ go build -o snakeai
$ ./snakeai -c
$ ./snakeai brain-0.json
```

Terminal-based Snake game

![scrrenshot](http://i.imgur.com/pHf4fjt.gif)

## Play

### Locally

```
$ go get github.com/DyegoCosta/snake-game
$ $GOPATH/bin/snake-game
```

### On Docker

```
$ docker run -ti dyego/snake-game
```

## Testing

```
$ cd $GOPATH/src/github.com/DyegoCosta/snake-game
$ make
```
