(define map (lambda (f x)
  (if (null? x)
    (quote ())
    (cons (f (car x)) (map f (cdr x))))))
