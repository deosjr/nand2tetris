(define map (lambda (f x)
  (if (null? x)
    (quote ())
    (cons (f (car x)) (map f (cdr x))))))

(define reverse (lambda (x) (begin
  (define reverse-acc (lambda (x acc)
    (if (null? x) acc
      (reverse-acc (cdr x) (cons (car x) acc)))))
  (reverse-acc x (quote ())))))

(define length (lambda (x)
  (if (null? x)
    0
    (+ 1 (length (cdr x))))))
