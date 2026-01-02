# python

## Basics

Run a script:  
`python app.py`

REPL:  
`python`

Create venv:  
`python -m venv .venv`

Activate venv:  
`source .venv/bin/activate`

Install packages:  
`pip install <pkg>`

Freeze deps:  
`pip freeze > requirements.txt`

## Syntax and types

Numbers/strings/bools:  
`1` `3.14` `"text"` `True` `False`

Lists/tuples/sets:  
`[1, 2]` `(1, 2)` `{1, 2}`

Dict:  
`{"a": 1, "b": 2}`

Unpack:  
`a, b = [1, 2]`  
`*rest, last = [1, 2, 3]`

## Functions

Define function:  
```py
def add(a, b=0):
    return a + b
```

Args and kwargs:  
```py
def f(*args, **kwargs):
    pass
```

Lambda:  
`key=lambda x: x["id"]`

## Control flow

If/elif/else:  
```py
if x > 0:
    pass
elif x == 0:
    pass
else:
    pass
```

For/while:  
```py
for item in items:
    pass
while condition:
    break
```

## Comprehensions

List/Dict/Set:  
`[x * 2 for x in xs if x > 0]`  
`{x: x * 2 for x in xs}`  
`{x for x in xs}`

Generator:  
`(x * 2 for x in xs)`

## Slicing

Slice list/string:  
`xs[1:5]` `xs[-1]` `xs[::2]`

## Strings

Format f-string:  
`f"{name} {count:02d}"`

Join/split:  
`",".join(parts)`  
`"a,b".split(",")`

## Files and paths

Read file:  
```py
with open("file.txt", "r", encoding="utf-8") as f:
    data = f.read()
```

Write file:  
```py
with open("file.txt", "w", encoding="utf-8") as f:
    f.write("hi")
```

Paths:  
```py
from pathlib import Path
Path("dir").mkdir(parents=True, exist_ok=True)
```

## Exceptions

Try/except/finally:  
```py
try:
    risky()
except ValueError as e:
    handle(e)
finally:
    cleanup()
```

## Classes and dataclasses

Class:  
```py
class User:
    def __init__(self, name):
        self.name = name
```

Dataclass:  
```py
from dataclasses import dataclass

@dataclass
class Point:
    x: int
    y: int
```

## Iterators and generators

Custom iterator:  
```py
class Counter:
    def __init__(self, n):
        self.n = n
    def __iter__(self):
        return self
    def __next__(self):
        if self.n <= 0:
            raise StopIteration
        self.n -= 1
        return self.n
```

Generator with yield:  
```py
def countdown(n):
    while n:
        yield n
        n -= 1
```

## Context managers

Custom context manager:  
```py
from contextlib import contextmanager

@contextmanager
def managed():
    setup()
    try:
        yield
    finally:
        teardown()
```

## Typing

Type hints:  
```py
from typing import Iterable, Optional

def total(xs: Iterable[int]) -> int:
    return sum(xs)
```

Literal and TypedDict:  
```py
from typing import Literal, TypedDict

class User(TypedDict):
    name: str
    role: Literal["admin", "user"]
```

## Async

Async/await:  
```py
import asyncio

async def main():
    await asyncio.sleep(1)

asyncio.run(main())
```

Gather tasks:  
`await asyncio.gather(task1(), task2())`

## Standard library highlights

JSON:  
`json.loads(text)` `json.dumps(obj)`

Datetime:  
`datetime.now()` `timedelta(days=1)`

Collections:  
`Counter` `defaultdict(list)` `deque`

## Packaging

Install in editable mode:  
`pip install -e .`

Build wheel:  
`python -m build`

## Testing

Pytest:  
`pytest -q`

Parametrize:  
```py
import pytest

@pytest.mark.parametrize("a,b", [(1, 2), (3, 4)])
def test_add(a, b):
    assert a + b > 0
```

## Data and scientific

NumPy array:  
`np.array([1, 2, 3])`

Pandas DataFrame:  
`pd.DataFrame({"a": [1, 2]})`

## Web

Requests:  
`requests.get("https://example.com")`

FastAPI route:  
```py
from fastapi import FastAPI
app = FastAPI()

@app.get("/ping")
def ping():
    return {"ok": True}
```

## Performance

Profile:  
`python -m cProfile app.py`

Timeit:  
`python -m timeit "sum(range(1000))"`
