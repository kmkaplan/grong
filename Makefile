include $(GOROOT)/src/Make.$(GOARCH)

TARBALL=/tmp/grong.tar.gz
# Distributed resolvers: as112.go, reflector-responder.go, rude-responder.go
RESPONDER=rude-responder.go

all: server

test: server
	./server -debug=4

server.$O: responder.$O types.$O

responder.$O: types.$O
	${GC} -o $@ $(RESPONDER)

%.$O: %.go 
	${GC} $<

server: server.$O
	${LD} -o $@ server.$O

dist: distclean
	(cd ..; tar czvf ${TARBALL} grong/*)

clean:
	rm -f server *.$O *.a

distclean: clean
	rm -f *~ responder.go