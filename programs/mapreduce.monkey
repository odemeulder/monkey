let map = fn(arr, func) {
  let iter = fn(arr, acc) {
    if (len(arr) == 0) {
      acc;
    } else {
      iter(rest(arr), push(acc, func(first(arr))));
    }
  };
  iter(arr, []);
};
let a = [1,2,3];
let square = fn(x) { x * x };
map(a,square);
let reduce = fn(arr, f, init) {
  let iter = fn(arr, acc) {
    if (len(arr) == 0) {
      acc
    } else {
      iter(rest(arr), f(acc, first(arr)))
    }
  }

  iter(arr, init);
}
let sum = fn(arr) { reduce(arr, fn(a,b) { a + b; }, 0) };
sum(a);