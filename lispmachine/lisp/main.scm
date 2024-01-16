(define sum-to (lambda (n)
  (if (= n 0)
      0
      (+ 1 (sum-to (- n 1))))))

(sum-to 500)
