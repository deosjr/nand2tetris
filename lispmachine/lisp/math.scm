#| wayy too slow
(define * (lambda (x y) (begin
    (define sum 0)
    (define sx x)
    (define mul-loop (lambda (i)
        (if (= i 13) sum
        (begin
        (if (bit y i)
          (set! sum (+ sum sx)) 0)
        (set! sx (+ sx sx))
        (mul-loop (+ i 1))
        ))))
    (mul-loop 0))))

(define * (lambda (x y)
    (if (= x 1)
      y
      (+ y (* (- x 1) y)))))
|#

#| 0x0001 = 0x4000 << 2|#
#|(define onebit (<< 0 2))|#

#|
(define bit (lambda (v index)
    (if (= index 0)
      (& v onebit)
      (& (<< v (- 16 index)) onebit))))
|#
#| 0x000f = (((0x4780 << 2) & 0x5e00) << 7) |#
(define last4mask (<< (& (<< 1920 2) 7680) 7))

(define / (lambda (x y)
    (if (> y x) 0
      ((lambda (twoq)
        (if (> y (- x (* y twoq)))
          twoq
          (+ 1 twoq)
       )) (times2 (/ x (times2 y))))
    )))

(define % (lambda (x n) (- x (* n (/ x n)))))

#| << is unsafe, so we need to put the number type prefix back! |#
(define times2 (lambda (x)
    (+ x x)))

