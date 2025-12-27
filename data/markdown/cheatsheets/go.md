# Go

## Version info and documentation

### Print the version of Go that is installed:

```
go version
```

### Get documentation on a package:

```
godoc fmt
```

### Get documentation on a function in a package:

```
godoc fmt Println
```


### Compile and run program:

```
go run main.go
```

## Language

### Types:

- `string`
- `int`
- `float32`
- `float64`
- `byte`

### Variable declaration:

```
var x string = ""
var x float64 = 3.1
x := "hi"
```

### Arrays:

```
var x [32]int
x := []float64{ 1, 2, 3, 4.5 }
```


### Slices:

```
var x[] int // empty slice
x := make([]float64, 5) // underlying array of length 5
x := arr[0:5]
x[:] // Slice of the whole array
slice2 := append(slice1, 4, 5)
copy(slice2, slice1)
x... (Slice which can be passed to variadic functions)
```

### Pointers:

```
x *int
// dereference
*x = 0
// address of
&x
xPtr := new(int). // Garbage collected, yolo
nil // for nullptr
```

### Structs:

```
type Circle struct {
  x float64
  y float64
  z float64
}
c := Circle{0, 0, 5}
c := Circle{x: 0, y: 0, r: 5}
c.r
```

### Methods w/ receivers:

```
func (c *Circle) area() float64 {
  return math.Pi * c.r * c.r
}

// Not inheritance but embedded types, can access Animal members on Cat
type Cat struct {
  Animal
  int lives
}
```

### Interfaces:

```
type Shape interface {
  area() float64
}
```

### Map:

```
x := make(map[string]int). // Map of a strings to ints
len(x) // Length of the map
delete(x)
elem, wasSuccess = x["hi"]
elements := map[string]string{
     "H": "Hydrogen",
     "He": "Helium",
}
```

### Conditional:

```
if i % 2 == 0 {
}

switch i {
case 0: fmt.Println("Zero")
case 1: fmt.Println("One")
case 2: fmt.Println("Two")
}
```

### Loops:

```
for i := 1; i < 10; i++ {
}
for i, value := range myArray {
}
for {
}
```


### Functions:

```
func main() {
}

func average(xs []float64) float64 {
  return 0;
}
// Can return multiple things
func f() (int, int) {
     return 5, 6
}
// Variadic
func add(args ...int) int {
}
// First class functions
add := func(x, y int) int {
  return x + y
}

// Defer function calls but make sure they happen
f, _ := os.Open(filename)
defer f.Close()

// panic / recover
divide := func(x, y) int {
  if y == 0 {
    panic("can’t divide by 0")
  }
  return x/ y
}
defer func() {
  str := recover()
  fmt.Println(str)
}()
```

### Casting:

```
float64(len(x))
```


### goroutines (coroutines):

```
go myFunc()
// Create a synchronous channel (reads / writes blocks)
var c chan string = make(chan string)
// Send data to a channel
c <- "hello"
// Get data from a channel
x := <- c
// Restrict to only send
func pinger(c chan<- string)
// Restrict to only receive
func reader(c <- chan string)
// Read from first available channel
select {
  case msg1 := <- c1:
    fmt.Println(msg1)
  case msg2 := <- c2:
    fmt.Println(msg2)
}
// Select and channels with a timeout
select {
case msg1 := <- c1:
     fmt.Println("Message 1", msg1)
case msg2 := <- c2:
     fmt.Println("Message 2", msg2)
case <- time.After(time.Second):
     fmt.Println("timeout")
// Select can also have a default: if nothing is ready
}

// Create an async buffered channel with capacity N, only waits if channel is full
 c := make(chan int, N)

```

### Packages and Imports:

```
import (
  "bytes"
  "container"
  "hash/crc32"
  "fmt"
  "time"
  "io"
  "os"
  "sort"
  "Strings"
  "math/rand"
)

import m "projects/myProject/math"
```

lowercase function names are not visible to the outside world. Uppercase are.

### Testing:

For file: `math.go`  
Name test: `math\_test.go`

To test function: `Average`  
Name function: `TestAverage`

