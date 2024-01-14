#| (map write (quote (1 2 3))) |#

(define sum-to (lambda (n)
  (if (= n 0)
      0
      (+ n (sum-to (- n 1))))))

(define sum2 (lambda (n acc)
  (if (= n 0)
      acc
      (sum2 (- n 1) (+ n acc)))))

(sum-to 10)

(sum2 10 0)
