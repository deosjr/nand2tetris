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
