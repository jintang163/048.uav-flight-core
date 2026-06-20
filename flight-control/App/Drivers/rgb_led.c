#include "rgb_led.h"
#include "main.h"
#include <math.h>

#define LED_TIMER_PERIOD 255

extern TIM_HandleTypeDef htim2;

static RGB_Color current_color;
static RGB_Color target_color;
static LedEffectType current_effect;
static bool led_on;
static uint32_t effect_start_time;
static uint32_t blink_period;
static float breathing_period;
static float rainbow_period;

#define BLINK_PERIOD_MS 500
#define BREATHING_PERIOD_MS 2000.0f
#define RAINBOW_PERIOD_MS 3000.0f

static void hsv_to_rgb(float h, float s, float v, uint8_t *r, uint8_t *g, uint8_t *b)
{
    float c = v * s;
    float x = c * (1.0f - fabsf(fmodf(h / 60.0f, 2.0f) - 1.0f));
    float m = v - c;

    float rf, gf, bf;

    if (h < 60.0f) {
        rf = c; gf = x; bf = 0;
    } else if (h < 120.0f) {
        rf = x; gf = c; bf = 0;
    } else if (h < 180.0f) {
        rf = 0; gf = c; bf = x;
    } else if (h < 240.0f) {
        rf = 0; gf = x; bf = c;
    } else if (h < 300.0f) {
        rf = x; gf = 0; bf = c;
    } else {
        rf = c; gf = 0; bf = x;
    }

    *r = (uint8_t)((rf + m) * 255.0f);
    *g = (uint8_t)((gf + m) * 255.0f);
    *b = (uint8_t)((bf + m) * 255.0f);
}

static void update_pwm(uint8_t r, uint8_t g, uint8_t b)
{
    __HAL_TIM_SET_COMPARE(&htim2, TIM_CHANNEL_1, r);
    __HAL_TIM_SET_COMPARE(&htim2, TIM_CHANNEL_2, g);
    __HAL_TIM_SET_COMPARE(&htim2, TIM_CHANNEL_3, b);
}

bool rgb_led_init(void)
{
    current_color.r = 0;
    current_color.g = 0;
    current_color.b = 0;
    target_color.r = 0;
    target_color.g = 0;
    target_color.b = 0;
    current_effect = LED_EFFECT_STATIC;
    led_on = false;
    effect_start_time = 0;
    blink_period = BLINK_PERIOD_MS;
    breathing_period = BREATHING_PERIOD_MS;
    rainbow_period = RAINBOW_PERIOD_MS;

    HAL_TIM_PWM_Start(&htim2, TIM_CHANNEL_1);
    HAL_TIM_PWM_Start(&htim2, TIM_CHANNEL_2);
    HAL_TIM_PWM_Start(&htim2, TIM_CHANNEL_3);

    update_pwm(0, 0, 0);

    return true;
}

void rgb_led_set_color(uint8_t r, uint8_t g, uint8_t b)
{
    target_color.r = r;
    target_color.g = g;
    target_color.b = b;
    current_effect = LED_EFFECT_STATIC;
    current_color = target_color;
    led_on = true;
    update_pwm(current_color.r, current_color.g, current_color.b);
}

void rgb_led_set_effect(LedEffectType effect, uint8_t r, uint8_t g, uint8_t b)
{
    target_color.r = r;
    target_color.g = g;
    target_color.b = b;
    current_effect = effect;
    effect_start_time = HAL_GetTick();
    led_on = true;
}

void rgb_led_update(uint32_t timestamp)
{
    if (!led_on) {
        update_pwm(0, 0, 0);
        return;
    }

    uint32_t elapsed = timestamp - effect_start_time;

    switch (current_effect) {
        case LED_EFFECT_STATIC:
            current_color = target_color;
            break;

        case LED_EFFECT_BLINK: {
            uint32_t phase = elapsed % blink_period;
            if (phase < blink_period / 2) {
                current_color = target_color;
            } else {
                current_color.r = 0;
                current_color.g = 0;
                current_color.b = 0;
            }
            break;
        }

        case LED_EFFECT_BREATHING: {
            float t = fmodf((float)elapsed / breathing_period, 1.0f);
            float brightness = (sinf(2.0f * M_PI * t - M_PI / 2.0f) + 1.0f) / 2.0f;
            current_color.r = (uint8_t)((float)target_color.r * brightness);
            current_color.g = (uint8_t)((float)target_color.g * brightness);
            current_color.b = (uint8_t)((float)target_color.b * brightness);
            break;
        }

        case LED_EFFECT_RAINBOW: {
            float hue = fmodf((float)elapsed / rainbow_period * 360.0f, 360.0f);
            hsv_to_rgb(hue, 1.0f, 1.0f, &current_color.r, &current_color.g, &current_color.b);
            break;
        }

        default:
            current_color = target_color;
            break;
    }

    update_pwm(current_color.r, current_color.g, current_color.b);
}

void rgb_led_off(void)
{
    led_on = false;
    update_pwm(0, 0, 0);
}

void rgb_led_on(void)
{
    led_on = true;
    effect_start_time = HAL_GetTick();
}
