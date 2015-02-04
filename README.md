# Let's play ground with Go

## Syntax

### Declaration

```
var foo int // declaration without initialization
var foo int = 42 // declaration with initialization
var foo, bar int = 42, 1302 // declare and init multiple vars at once
var foo = 42 // type omitted, will be inferred
foo := 42 // shorthand, only in func bodies, omit var keyword, type is always implicit
const constant = "This is a constant"
```


### function

```
// a simple function
func functionName() {}

// function with parameters (again, types go after identifiers)
func functionName(param1 string, param2 int) {}

// multiple parameters of the same type
func functionName(param1, param2 int) {}

// return type declaration
func functionName() int {
     return 42
}

// Can return multiple values at once
func returnMulti() (int, string) {
     return 42, "foobar"
}
var x, str = returnMulti()

// Return multiple named results simply by return
func returnMulti2() (n int, s string) {
     n = 42
     s = "foobar"
     // n and s will be returned
     return
}
var x, str = returnMulti2()
```


### Functions As Values And Closures

```
func main() {
     // assign a function to a name
     add := func(a, b int) int {
          return a + b
     }
     // use the name to call the function
     fmt.Println(add(3, 4))
}

// Closures: Functions can access values that were in scope when defining the
// function

// adder returns an anonymous function with a closure containing the variable sum
func adder() func(int) int {
     sum := 0
     return func(x int) int {
          sum += x // sum is declared outside, but still visible
          return sum
     }
}
```

### Loops

```
// There's only `for`, no `while`, no `until`
for i := 1; i < 10; i++ {
}

for ; i < 10; { // while - loop
}

for i < 10 { // you can omit semicolons if there is only a condition
}

for { // you can omit the condition ~ while (true)
}
```

### controls

```
// for 에서처럼 if 에서 := 를 사용하는것은 y에 먼저 값을 대입하고,
// 그리고 y > x를 검사한다는 의미.
if y := expensiveComputation(); y > x {
     x = y
}
```


### defer, panic, recover

```
func CopyFile(dstName, srcName string) (written int64, err error) {
     src, err := os.Open(srcName)
     if err != nil {
          return
     }
     defer src.Close()

     dst, err := os.Create(dstName)
     if err != nil {
          return
     }
     defer dst.Close()

     return io.Copy(dst, src)
}
```

1. A deferred function's arguments are evaluated when the defer statement is evaluated.
2. Deferred function calls are executed in Last In First Out order after_the surrounding function returns.
3. Deferred functions may read and assign to the returning function's named return values.

http://blog.golang.org/defer-panic-and-recover

## 더 볼것

- struct, empty struct, interface, inheritance
- defer, panic, recover. excetpion 핸들링은 ???
- concurrency. go routine은 뭐야
- channel. another smalltalk ?

## good parts

```
if _, err := strconv.Atoi("non-int"); err != nil {
    fmt.Println(err)
}

# to be continued
