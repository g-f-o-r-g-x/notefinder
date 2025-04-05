AC_INIT([notefinder], [0.1])
AC_CONFIG_SRCDIR([configure.ac])
AC_CONFIG_HEADERS([config.h])
AC_PROG_CC

AC_PATH_PROG([GO], [go], [no])
AS_IF([test "x$GO" = "xno"], [AC_MSG_ERROR([go compiler not found])])

AC_SUBST([CGO_CFLAGS], [`perl -MExtUtils::Embed -e ccopts | sed -e 's/-D_GNU_SOURCE//g'`])
AC_SUBST([CGO_LDFLAGS], [`perl -MExtUtils::Embed -e ldopts`])

PERL_VENDORLIB=`perl -MConfig -e 'print $Config{installvendorlib}'`
AC_SUBST([PERL_VENDORLIB])

AC_CONFIG_FILES([Makefile])
AC_OUTPUT
