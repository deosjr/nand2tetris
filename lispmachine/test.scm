(begin
    (define x 7)
    (define y 8)
    (define z 9)
)

(quote 42)

(begin x y z)

(quote z)

(define test (lambda (x) x))

(test 42)

(map test (quote (1 2 3)))
(map write (quote (1 2 3)))

(if (> 3 x) (- 1 2) 42)

(if (> 3 x) fail 42)

(if (> 3 fail) (- 1 2) 42)
