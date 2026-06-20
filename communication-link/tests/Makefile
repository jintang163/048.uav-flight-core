CC = gcc
CFLAGS = -Wall -Wextra -std=c11 -g
INCLUDES = -I../include -I../mavlink -I../crypto -I../transport
LDFLAGS = -L.. -lcommunication_link -lpthread

TESTS = test_mavlink_parser test_sm4 test_transport

all: $(TESTS)

test_mavlink_parser: test_mavlink_parser.c
	$(CC) $(CFLAGS) $(INCLUDES) $< -o $@ $(LDFLAGS)

test_sm4: test_sm4.c
	$(CC) $(CFLAGS) $(INCLUDES) $< -o $@ $(LDFLAGS)

test_transport: test_transport.c
	$(CC) $(CFLAGS) $(INCLUDES) $< -o $@ $(LDFLAGS)

clean:
	rm -f $(TESTS) *.o

.PHONY: all clean
