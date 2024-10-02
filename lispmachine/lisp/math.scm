(define * (lambda (x y)
    (if (= x 1)
      y
      (+ y (* (- x 1) y)))))

(define / (lambda (x y)
    (if (> y x) 0
      ((lambda (twoq)
        (if (> y (- x (* y twoq)))
          twoq
          (+ 1 twoq)
       )) (<< (/ x (<< y 1)) 1))
    )))

(define % (lambda (x n) (- x (* n (/ x n)))))
