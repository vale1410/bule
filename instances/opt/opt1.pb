* #variable= 2 #constraint= 4
****************************************
* begin normalizer comments
* category= OPT-SMALLINT
* end normalizer comments
****************************************
min: -1 x1 ;
+2 x1 +2 x2 >= 1 ;
-2 x1 -2 x2 >= -3 ;
-2 x1 +2 x2 >= -1 ;
+2 x1 -2 x2 >= -1 ;
