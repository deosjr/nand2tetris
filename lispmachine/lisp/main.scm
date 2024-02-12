#| TODO: trie with value insert index |#
(define symbol-table (quote (
    (#\i #\f)
    (#\d #\e #\f #\i #\n #\e)
    (#\q #\u #\o #\t #\e)
    (#\s #\e #\t #\!)
    (#\l #\a #\m #\b #\d #\a)
    (#\b #\e #\g #\i #\n)
)))

(define read-token (lambda ()
    (begin 
      (consume-whitespace)
      (read-until-token (quote ()))
    )))

(define consume-whitespace (lambda ()
    #| (let ((peeked (peek-char))) .. |#
    ((lambda (peeked)
       #| eq? doesnt exist, only = |#
       (if (= peeked #\space)
         #| todo: if without alt! |#
         (begin (read-char) (consume-whitespace)) 0
    )) (peek-char))))

#| assumes whitespace has been consumed |#
(define read-until-token (lambda (stack)
    #| (let ((peeked (peek-char))) .. |#
    ((lambda (peeked)
       (if (or (= peeked #\eof) (= peeked #\space))
         (reverse stack)
         #| todo: parse escaped bracket chars |#
         (if (or (= peeked 40) (= peeked 41))
           (if (null? stack)
             (cons (read-char) (quote ()))
             (reverse stack))
           (read-until-token (cons (read-char) stack))
    ))) (peek-char))))

(define newline (lambda () (write-char #\newline)))

(begin
  (map write-char (read-token))
  (newline)
  (map write-char (read-token))
  (newline)
  (map write-char (read-token))
  (newline)
  (map write-char (read-token))
  (newline)
)
