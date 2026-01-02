# c++

## Basics

Compile and run:  
`g++ -std=c++20 -O2 -Wall main.cpp -o app`

Run app:  
`./app`

Include header:  
`#include <vector>`

Import module:  
`import std;`

Header unit (compiler support):  
`import <vector>;`

## Types and literals

Fixed width ints:  
`int32_t` `uint64_t`

Auto type:  
`auto x = 42;`

Uniform init:  
`std::vector<int> v{1, 2, 3};`

## References and pointers

Reference:  
`int& r = x;`

Const reference:  
`const std::string& s = name;`

Smart pointers:  
`std::unique_ptr<T>` exclusive ownership  
`std::shared_ptr<T>` shared ownership  
`std::weak_ptr<T>` non-owning observer

## RAII

Scope cleanup:  
```cpp
std::lock_guard<std::mutex> lock(mu);
```

## Functions

Default args:  
`int add(int a, int b = 0);`

Overload:  
`void f(int);` `void f(std::string_view);`

Lambda:  
`auto f = [](int x) { return x * 2; };`

## Classes

Rule of 0:  
`use defaults + RAII types`

Rule of 5:  
`copy/move ctor, copy/move assign, dtor`

Defaulted/deleted:  
`Foo() = default;` `Foo(const Foo&) = delete;`

## Move semantics

Move:  
`std::move(obj)`

Perfect forwarding:  
```cpp
template <class T>
void f(T&& x) { g(std::forward<T>(x)); }
```

## const and constexpr

Constexpr function:  
```cpp
constexpr int add(int a, int b) { return a + b; }
```

Consteval:  
`consteval int id() { return 42; }`

## Ranges (C++20)

Range algorithms:  
```cpp
std::ranges::sort(v);
```

Views:  
```cpp
auto odds = v | std::views::filter([](int x){ return x % 2; });
```

## Concepts (C++20)

Requires clause:  
```cpp
template <typename T>
requires std::integral<T>
T add(T a, T b) { return a + b; }
```

## std::optional and std::variant

Optional:  
`std::optional<int> x;`

Variant:  
`std::variant<int, std::string> v;`

Visit:  
```cpp
std::visit([](auto&& val){}, v);
```

## std::string_view

Non-owning view:  
`std::string_view sv = "text";`

## Error handling

Exceptions:  
`throw std::runtime_error("fail");`

Expected (C++23):  
`std::expected<T, E>`

## Concurrency

Thread:  
`std::thread t(fn);`

Future/async:  
`auto f = std::async(fn);`

Atomic:  
`std::atomic<int> x{0};`

## Modules (C++20)

Module interface:  
```cpp
export module math;
export int add(int a, int b);
```

Import:  
`import math;`

## Coroutines (C++20)

Generator type (sketch):  
```cpp
// requires a generator type implementation
co_yield value;
```

## Filesystem

Path and exists:  
```cpp
std::filesystem::path p = "file.txt";
std::filesystem::exists(p);
```

## Formatting (C++20)

Format:  
`std::format("hello {}", name);`

## Testing and tooling

Sanitizers:  
`-fsanitize=address,undefined`

Clang-Tidy:  
`clang-tidy main.cpp -- -std=c++20`

## Build systems

CMake minimal:  
```cmake
cmake_minimum_required(VERSION 3.20)
project(app)
add_executable(app main.cpp)
```
