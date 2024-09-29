(define length (lambda (x) (if (null? x) 0 (+ 1 (length (cdr x))))))
