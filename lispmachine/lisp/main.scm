#|(map write (quote (#\H #\e #\l #\l #\o #\space #\W #\o #\r #\l #\d #\!))) |#

(map (lambda (x) (write (+ x 1))) (quote (1 2 3)))
