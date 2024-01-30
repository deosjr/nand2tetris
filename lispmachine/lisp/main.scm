(define sum-to (lambda (n)
  (if (= n 0)
      0
      (+ 1 (sum-to (- n 1))))))

(define sum2 (lambda (n acc)
  (if (= n 0)
      acc
      (sum2 (- n 1) (+ 1 acc)))))

#| 480k ticks unoptimised|#
#| 430k ticks with eval taking 2 args |#
#| 220k ish on firstpass handcrafting asm |#
#| 200k when hardcoding special lookup |#
(sum2 50 0)
