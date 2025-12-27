# Rust


## Install and update

### Install Rustup & Rust:

```
curl https://sh.rustup.rs -sSf | sh
```

### Update Rust:

```
rustup update
```

## Documentation:

### Rust documentation 
```
rustup doc
```

### Cargo documentation

https://doc.rust-lang.org/cargo/.

### Per project documentation

```
cargo doc --open
```

### Formatting tool:

```
cargo fmt -- --check
```

### New project:

```
cargo new hello_cargo --bin
```

### Build:

```
cargo build
cargo build --release
```

### Build & Run:

```
cargo run
cargo run --release
```

### Updating crates including Cargo.lock file:

cargo update

### Checking for errors when building

Check for errors without building:
cargo check


## Language

### Printing text:

```
println!("You guessed: {}", guess);
```

### Bring a library into scope:

```
use std::io;
```

### Creating a variable:

```
let foo = bar; // immutable
Let mut foo = String::new(); // mutable
```

### Shadowing:

You can shadow a variable to avoid having to create another one.

```
let mut guess = String::new()
let guess : u32 = guess.trim().parse().expect("enter a number!");
```

### Reference:

```
&foo // immutable reference
&mut foo // mutable reference
```

### Macros:

Calling a macro (like println!) you put an exclamation mark at the end.

### Statements and expressions:

Statements end in a semicolon don't return a value
Expressions have no semicolon and return a value.

### Statement: 

```
let x = 3;
```

### Expressions: 

```
{ let x = 3; x +1 } // Returns x+1
if (x == 3) { 5 } else { 3 }
```

## Config

### Using a crate:

- Packages in Rust are called crates.
- Specify the dependency in the Cargo.toml file, eg: rand = "0.3.14"
- Reference it at the top of your file: `extern crate rand;` and use `rand::Rng`
