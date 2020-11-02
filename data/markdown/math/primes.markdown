# Mathematics - Primes

## What are prime numbers?

A natural number is a number in the set {1, 2, 3, ...}. Prime numbers are a subset of natural numbers.

In particular prime numbers are the subset of natural numbers more than 1 and who have no divisors other than themselves and 1.

4 is not a prime because 4 can be factored into 2 * 2.
5 is a prime because no numbers multiplied by each other give 5 other than 1 and 5 itself.

## How many primes are there?

There are infinitely many prime numbers.

How can you be sure? Well assume you have a set of all the finite number of primes:
P = {p1, p2, ... pN}.

Then multiply all of those numbers together and add one. n = p1 * p2 * ... * pN + 1.

We know that p does not divide n for each p in the set P because n mod p == 1.

But the Fundamental Theorem of Algebra states that every natural number is itself a prime or can be factored into primes in a unique way.

Since we determined that N can't be factored using any prime in our set P, then it must itself be a prime or can be factored into different primes not completely in our set.

Therefore for any set of primes, we can always use that set to generate a bigger set of primes. We would generate this larger set, by multiplying the entire set together, and adding 1. Then we'd have a number which is prime or which contains prime factors not in our set.

And since we can always keep generating larger sets of primes infinitely; therefore, there are infinitely many primes.

## How to verify if a number is prime or not?

The simplest way to verify if a number n is prime is to check if any number that is greater than 1 and smaller than the square root of n is a divisor of n. More specifically you only need to see if the primes smaller than the square root of n are divisors.

But the simplest method of verifying primality becomes the hardest method when we have large numbers we want to verify because the number of factors we have to check grow fast.

There are 2 types of primality tests:

- Deterministic: Determines for sure if a number is prime or not
- Probabilistic: Determines for sure if a number is prime, but may report a composite to be prime

So why would you ever use a probabilistic test when a deterministic test is available?

Because probabilistic tests can be much faster to run. So if you're searching for a prime, you can use the probabilistic test to find a number that is probably a prime, and then fall back to a deterministic test to be sure for the numbers the probabilistic test deems to be probably a primes.

Some examples of other primality tests are: Fermat's primality test, Elliptic curve primality test, and AKS primality test. There are many others though.

## How common are prime numbers in the set of natural numbers?
We know that there are infinitely many primes, but how common are they in comparison to composite (non prime) numbers?

The distribution of primes within the natural numbers is not known even to the greatest minds in mathematics throughout history. There are laws that govern the average distribution such as the prime number theorem. But it seems as though we may never understand the distribution of primes fully.

The prime number theorem states that if you select any number N, and then take it's number of digits, then the chance of it being a prime is only 1/N. So as you select bigger numbers for N, you have less and less of a chance of them being primes.
