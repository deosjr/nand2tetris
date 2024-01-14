(define map (lambda (f x)
                (if (null? x) (quote ())
                  (cons (f (car x)) (map f (cdr x))))))

(null? (quote ()))

(null? 1)

(car (cdr (quote (1 2 3))))

(map (lambda (x) (write (+ x 1))) (quote (1 2 3)))
