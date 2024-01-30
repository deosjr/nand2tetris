(define garbage (lambda ()
    (begin
        (cons 1 1)
        (garbage))))

(garbage)
