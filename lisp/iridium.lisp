(eval-when (:compile-toplevel)
  (ql:quickload "cl-utilities")
  )

(require "cl-utilities")

(defpackage :iridium
  (:use :common-lisp :cl-utilities)
  (:export ))

;; (in-package :iridium)

;; Annoyance: the path needs to end with a slash to be interpreted as a directory.
(defparameter *game-dir* #P"/Users/riccardo/git/iridium/ignore/testcl/")

(defconstant +max-width+ 78)

(defun save-config (config dir)
  "Save a configuration file to the game directory"
  (let ((filepath (merge-pathnames "game.sxp" dir)))
    (with-open-file (out filepath :direction :output :if-exists :supersede)
      (with-standard-io-syntax
        (print config out)))))

(defun load-config (dir)
  "Load a configuration file from the game directory"
  (let ((filepath (merge-pathnames "game.sxp" dir )))
    (with-open-file (in filepath)
      (with-standard-io-syntax
        (read in)))))

(defun print-title (config width)
  (let ((line (make-string width :initial-element #\-)))
    (format t "~a~%~%" line)
    (format t "~v:@<~A~>~%" width (string-upcase (getf config :title)))
    (when (getf config :subtitle)
      (format t "~%~v:@<~a~>~%~%" width (getf config :subtitle)))
    (format t "~%~v:@<By ~a~>~%" width (getf config :author))
    (format t "~%~a~%" line)))

(defun run (dir)
  "Run a game in the CLI"
  (let ((config (load-config dir))
        (width (compute-terminal-width)))
    (print-title config width)
    (main-loop dir (getf config :init) width)))

(defun compute-terminal-width ()
  (let ((cols (sb-ext:posix-getenv "COLUMNS")))
    (if (equal cols "")
        +max-width+
        ;; Error checking?
        (multiple-value-bind (i) (parse-integer cols) (- i 2)))))

(defun main-loop (dir passage-name width)
  (loop (let ((passage (read-passage dir passage-name)))
          (format t "~%")
          (unless passage
            (format t "Passage name ~w does not exist~%~%" passage-name)
            (return nil))
          (destructuring-bind (paragraphs options) (process-passage passage)
            (dolist (p paragraphs)
              (print-paragraph p width))
            (when (not options)
              (return t))
            (loop for opt in options
                  for index from 1
                  do (format t " ~2d. " index)
                  do (print-text (cadr opt) width))
            (format t "~%")
            (let ((next-passage-name (read-choice options)))
              ;; TODO: allow quitting.
              (if next-passage-name
                  (setq passage-name next-passage-name)
                  (return t)))))))

(defun read-choice (options)
  (loop (let ((r (progn (format t "? ") (finish-output) (read))))
          (when (equal r 'q)
            (return nil))
          (when (and (integerp r)
                     (> r 0)
                     (<= r (length options)))
            (return (car (nth (- r 1) options)))))))

(defun read-passage (dir name)
  "Read a passage from the game directory"
  (let ((filepath (merge-pathnames (format nil "~a.txt" name) (merge-pathnames #P"passages/" dir))))
    (with-open-file (in filepath :if-does-not-exist nil)
      (when in 
        (with-standard-io-syntax
          (loop for content = (read in nil nil)
                while content
                collect content))))))

(defun process-passage (passage)
  (let ((options nil)
        (text nil))
    (dolist (item passage)
      (case (car item)
        (p (push (cdr item) text))
        (option (push (cdr item) options))))
    (list (nreverse text) (nreverse options))))

(defvar *emit-width* +max-width+)
(defvar *emit-indent1* 0)
(defvar *emit-indent* 0)
(defvar *emit-words* nil)
(defvar *emit-current-word* "")

(defun emit-start (indent1 indent width)
  (setq *emit-width* width)
  (setq *emit-indent1* indent1)
  (setq *emit-indent* indent)
  (setq *emit-words* nil)
  (setq *emit-current-word* ""))

(defun emit-word (word)
  (setq *emit-current-word* (concatenate 'string *emit-current-word* word)))

(defun emit-space ()
  (unless (string= *emit-current-word* "")
    (push *emit-current-word* *emit-words*)
    (setq *emit-current-word* "")))

(defun emit-end ()
  (emit-space)
  (let ((curr *emit-indent1*)
        (width *emit-width*))
    (format t "~a" (make-string *emit-indent1* :initial-element #\Space))
    (dolist (w (nreverse *emit-words*))
      (when (> (+ curr (length w)) width)
        (format t "~%")
        (format t "~a" (make-string *emit-indent* :initial-element #\Space))
        (setq curr *emit-indent*))
      (format t "~a " w)
      (setq curr (+ curr (length w) 1)))
    (unless (= curr 0)
      (format t "~%"))))

(defun emit-entries (entries)
  (dolist (entry entries)
    (cond ((stringp entry)
           (let ((words (uiop:split-string entry :separator '(#\Space #\Tab #\Newline))))
             (unless (null words)
               (emit-word (car words))
               (dolist (w (cdr words))
                 (emit-space)
                 (emit-word w)))))
          ((and (consp entry) (equal (car entry) 'dialog))
           (emit-word "“")
           (emit-entries (cdr entry))
           (emit-word "”"))
          (t (format t "((unknown form: ~a)) " entry)))))
            
(defun print-text (text width)
  (emit-start 0 5 width)
  (emit-entries (list text))
  (emit-end))
  
(defun print-paragraph (para width)
  (emit-start 0 0 width)
  (emit-entries para)
  (emit-end)
  (format t "~%"))


;; Better way of printing?
;;
;;   *emit-width*
;;   *emit-indent1*    ;; indent of first line
;;   *emit-indent*     ;; indent of subsequent lines
;;   *emit-words*
;;   *emit-current-word*
;;
;;   (emit-start indent1 indent width)
;;
;;   (emit-word word)    ;; add to current word
;; 
;;   (emit-space)         ;; move current word to words if not empty
;;   (emit-end)            ;; output stuff in the given width and a newline at the end
;;
;; emit accumulates all the outputs until emit-end, when the "formatting" and output is done
;; 
