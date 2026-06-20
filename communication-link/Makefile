CC = gcc
AR = ar
CFLAGS = -Wall -Wextra -Werror -std=c11 -fPIC
LDFLAGS = -lpthread

ifeq ($(DEBUG),1)
CFLAGS += -g -O0 -DDEBUG
else
CFLAGS += -O2
endif

INCLUDES = -I./include -I./mavlink -I./crypto -I./transport

MAVLINK_SRC = mavlink/mavlink_v2.c mavlink/parser.c mavlink/encoder.c \
              mavlink/heartbeat.c mavlink/command.c mavlink/mission.c \
              mavlink/parameter.c

CRYPTO_SRC = crypto/sm4.c crypto/tls_manager.c crypto/crypto_manager.c \
             crypto/key_exchange.c crypto/auth.c

TRANSPORT_SRC = transport/transport.c transport/uart_transport.c \
                transport/tcp_transport.c transport/udp_transport.c \
                transport/transport_mux.c transport/link_monitor.c \
                transport/fragment.c

SRCS = $(MAVLINK_SRC) $(CRYPTO_SRC) $(TRANSPORT_SRC)
OBJS = $(SRCS:.c=.o)

TARGET = libcommunication_link.a

all: $(TARGET) tests examples

$(TARGET): $(OBJS)
	$(AR) rcs $@ $(OBJS)

%.o: %.c
	$(CC) $(CFLAGS) $(INCLUDES) -c $< -o $@

tests: $(TARGET)
	$(MAKE) -C tests

examples: $(TARGET)
	$(MAKE) -C examples

clean:
	rm -f $(OBJS) $(TARGET)
	$(MAKE) -C tests clean
	$(MAKE) -C examples clean

install: $(TARGET)
	install -d $(PREFIX)/lib
	install -m 644 $(TARGET) $(PREFIX)/lib
	install -d $(PREFIX)/include/communication_link
	install -m 644 include/*.h $(PREFIX)/include/communication_link
	install -d $(PREFIX)/include/communication_link/mavlink
	install -m 644 mavlink/*.h $(PREFIX)/include/communication_link/mavlink
	install -d $(PREFIX)/include/communication_link/crypto
	install -m 644 crypto/*.h $(PREFIX)/include/communication_link/crypto
	install -d $(PREFIX)/include/communication_link/transport
	install -m 644 transport/*.h $(PREFIX)/include/communication_link/transport

.PHONY: all tests examples clean install
