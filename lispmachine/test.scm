(begin
    (define x 7)
    (define y 8)
    (define z 9)
)

(quote 42)

(begin x y z)

(quote z)

(define test (lambda (a b) (+ a b)))

(test 1 41)

(map write (quote (1 2 3)))

(apply test (quote (41 1)))

(if (> 3 x) (- 1 2) 42)

(if (> 3 x) fail 42)

(if (> 3 fail) (- 1 2) 42)
