(define sum-to (lambda (n)
  (if (= n 0)
      0
      (+ 1 (sum-to (- n 1))))))

(define sum2 (lambda (n acc)
  (if (= n 0)
      acc
      (sum2 (- n 1) (+ 1 acc)))))

(sum2 500 0)
