#ifndef __RGB_LED_H__
#define __RGB_LED_H__

#include "types.h"
#include "flight_config.h"

typedef enum {
    LED_EFFECT_STATIC = 0,
    LED_EFFECT_BLINK = 1,
    LED_EFFECT_RAINBOW = 2,
    LED_EFFECT_BREATHING = 3
} LedEffectType;

typedef struct {
    uint8_t r;
    uint8_t g;
    uint8_t b;
} RGB_Color;

bool rgb_led_init(void);
void rgb_led_set_color(uint8_t r, uint8_t g, uint8_t b);
void rgb_led_set_effect(LedEffectType effect, uint8_t r, uint8_t g, uint8_t b);
void rgb_led_update(uint32_t timestamp);
void rgb_led_off(void);
void rgb_led_on(void);

#endif
