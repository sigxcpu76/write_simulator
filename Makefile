OBJ:=write_simulator

all: $(OBJ)

$(OBJ): *.go
	go build --ldflags '-extldflags "-static" -w -s'

clean:
	rm -f $(OBJ)