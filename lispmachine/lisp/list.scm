(define map (lambda (f x)
  (if (null? x)
    (quote ())
    (cons (f (car x)) (map f (cdr x))))))

(define reverse (lambda (x) (reverse-acc x (quote ()))))
(define reverse-acc (lambda (x acc)
    (if (null? x) acc
      (reverse-acc (cdr x) (cons (car x) acc)))))
