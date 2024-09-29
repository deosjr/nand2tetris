#| TODO: trie with value insert index |#
(define symbol-table (quote ()))
(define symbol-table-size 0)

(define add-symbol (lambda (symbol)
    (begin
      (set! symbol-table (cons (cons symbol symbol-table-size) symbol-table))
      (set! symbol-table-size (+ symbol-table-size 1))
    )))

(add-symbol (quote (#\i #\f)))
(add-symbol (quote (#\d #\e #\f #\i #\n #\e)))
(add-symbol (quote (#\q #\u #\o #\t #\e)))
(add-symbol (quote (#\s #\e #\t #\!)))
(add-symbol (quote (#\l #\a #\m #\b #\d #\a)))
(add-symbol (quote (#\b #\e #\g #\i #\n)))

(define get-symbol (lambda (symbol) (begin
    (define get (lambda (symbol table)
        (if (null? table) (quote #f)
          (if (string-eq? (car (car table)) symbol)
            (cdr (car table))
            (get symbol (cdr table))
        ))))
    (get symbol symbol-table))))

(define read-token (lambda ()
    (begin 
      (consume-whitespace)
      (read-until-token (quote ()))
    )))

(define consume-whitespace (lambda ()
    #| (let ((peeked (peek-char))) .. |#
    ((lambda (peeked)
       #| eq? doesnt exist, only = |#
       (if (or (= peeked #\space) (= peeked #\newline))
         #| todo: if without alt! |#
         (begin (read-char) (consume-whitespace)) 0
    )) (peek-char))))

#| assumes whitespace has been consumed |#
(define read-until-token (lambda (stack)
    #| (let ((peeked (peek-char))) .. |#
    ((lambda (peeked)
       (if (or (or (= peeked #\eof) (= peeked #\space)) (= peeked #\newline))
         (reverse stack)
         #| todo: parse escaped bracket chars |#
         (if (or (= peeked 40) (= peeked 41))
           (if (null? stack)
             (cons (read-char) (quote ()))
             (reverse stack))
           (read-until-token (cons (read-char) stack))
    ))) (peek-char))))

(define char-isdigit? (lambda (c)
  (and (> c 47) (> 58 c))))

(define char->digit (lambda (c)
  (if (char-isdigit? c) (- c 48) (quote #f))))

#| returns modified stack |#
(define make-list (lambda (stack) (begin
  (define list-rec (lambda (acc stack)
    (if (null? stack)
      (error 45) #| todo: parselist error |#
      ((lambda (p)
         (if (= p 40)
           (cons acc stack)
           (list-rec (cons p acc) (cdr stack))
      )) (car stack)))))
  (list-rec (quote ()) stack))))

(define make-atom (lambda (token)
  (if (char-isdigit? (car token))
    (make-primitive token 0)
    (make-symbol token))))

(define make-primitive (lambda (token acc)
    ((lambda (c t)
      (if (null? token)
        acc
      (if (char-isdigit? c)
        (make-primitive t (+ (* 10 acc) (char->digit c)))
        (error 42) #| todo: parsenum error |#
      ))) (car token) (cdr token))))

#| TODO: still a number under the hood atm |#
(define make-symbol (lambda (token)
    #| (let ((got (get-symbol token))) .. |#
    ((lambda (got)
       (if got got
         (begin
           (add-symbol token)
           (- symbol-table-size 1))))
                    (get-symbol token))))

(define read-file (lambda () (begin
  (define read-file-rec (lambda (stack)
    #| (let ((token (read-token))) .. |#
    ((lambda (token)
       (if (null? token)
         #\eof
         #| (let ((len (length token))) .. |#
         ((lambda (len)
            #| token == '(' |#
            (if (and (= 1 len) (= (car token) 40))
              (read-file-rec (cons 40 stack))
            #| token == ')' |#
            (if (and (= 1 len) (= (car token) 41))
              (read-file-rec (make-list stack))
            #| else |#
              (read-file-rec (cons (make-atom token) stack))
         ))) (length token))
     )) (read-token))))
  (read-file-rec (quote ())))))

(define newline (lambda () (write-char #\newline)))

(define * (lambda (x y)
    (if (= x 1)
      y
      (+ y (* (- x 1) y)))))

(define string-eq? (lambda (x y)
    (if (null? x) (null? y)
      (if (null? y) (quote #f) #| why does #f not work here? |#
        (if (= (car x) (car y))
          (string-eq? (cdr x) (cdr y))
          (quote #f))))))

(read-file)
