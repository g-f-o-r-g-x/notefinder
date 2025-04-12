NAME = notefinder

prefix = /usr
bindir = ${exec_prefix}/bin
exec_prefix = ${prefix}

CGO_CFLAGS =  -D_REENTRANT  -fwrapv -fno-strict-aliasing -pipe -fstack-protector-strong -D_LARGEFILE_SOURCE -D_FILE_OFFSET_BITS=64  -I/usr/lib/perl5/5.40.0/x86_64-linux-thread-multi/CORE 
CGO_LDFLAGS = -Wl,-E -Wl,-rpath,/usr/lib/perl5/5.40.0/x86_64-linux-thread-multi/CORE  -L/usr/local/lib64 -fstack-protector-strong  -L/usr/lib/perl5/5.40.0/x86_64-linux-thread-multi/CORE -lperl -lm -ldl -lcrypt -lpthread

PERL_VENDORLIB = /usr/lib/perl5/vendor_perl/5.40.0

all: compile

compile:
	CGO_CFLAGS="$(CGO_CFLAGS) -Wno-builtin-macro-redefined" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	go build -o $(NAME) .

install: compile
	mkdir -p $(DESTDIR)$(bindir)
	cp ./$(NAME) $(DESTDIR)$(bindir)/$(NAME)
	mkdir -p $(DESTDIR)$(PERL_VENDORLIB)
	cp ./Notefinder.pm $(DESTDIR)$(PERL_VENDORLIB)/Notefinder.pm
