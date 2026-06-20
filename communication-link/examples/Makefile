CC = gcc
CFLAGS = -Wall -Wextra -std=c11 -O2
INCLUDES = -I../include -I../mavlink -I../crypto -I../transport
LDFLAGS = -L.. -lcommunication_link -lpthread

EXAMPLES = example_uav_gcs example_secure_link example_multi_link

all: $(EXAMPLES)

example_uav_gcs: example_uav_gcs.c
	$(CC) $(CFLAGS) $(INCLUDES) $< -o $@ $(LDFLAGS)

example_secure_link: example_secure_link.c
	$(CC) $(CFLAGS) $(INCLUDES) $< -o $@ $(LDFLAGS)

example_multi_link: example_multi_link.c
	$(CC) $(CFLAGS) $(INCLUDES) $< -o $@ $(LDFLAGS)

clean:
	rm -f $(EXAMPLES) *.o

.PHONY: all clean
