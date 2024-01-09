(begin
    (define x 7)
    (define y 8)
    (define z 9)
)

(begin x y z)

(quote z)

(if (> 3 x) (- 1 2) 42)

(if (> 3 x) fail 42)

(if (> 3 fail) (- 1 2) 42)
