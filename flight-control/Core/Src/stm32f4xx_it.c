#include "main.h"
#include "stm32f4xx_it.h"

extern TIM_HandleTypeDef htim1;
extern UART_HandleTypeDef huart1;
extern UART_HandleTypeDef huart2;
extern UART_HandleTypeDef huart3;
extern UART_HandleTypeDef huart4;
extern DMA_HandleTypeDef hdma_usart1_rx;
extern DMA_HandleTypeDef hdma_usart2_rx;
extern DMA_HandleTypeDef hdma_usart3_rx;
extern DMA_HandleTypeDef hdma_uart4_rx;

void NMI_Handler(void)
{
    while (1) {
    }
}

void HardFault_Handler(void)
{
    while (1) {
    }
}

void MemManage_Handler(void)
{
    while (1) {
    }
}

void BusFault_Handler(void)
{
    while (1) {
    }
}

void UsageFault_Handler(void)
{
    while (1) {
    }
}

void SVC_Handler(void)
{
}

void DebugMon_Handler(void)
{
}

void PendSV_Handler(void)
{
}

void SysTick_Handler(void)
{
    HAL_IncTick();
}

void DMA1_Stream5_IRQHandler(void)
{
    HAL_DMA_IRQHandler(&hdma_usart2_rx);
}

void DMA1_Stream6_IRQHandler(void)
{
    HAL_DMA_IRQHandler(&hdma_usart3_rx);
}

void DMA2_Stream2_IRQHandler(void)
{
    HAL_DMA_IRQHandler(&hdma_usart1_rx);
}

void USART1_IRQHandler(void)
{
    HAL_UART_IRQHandler(&huart1);
}

void USART2_IRQHandler(void)
{
    HAL_UART_IRQHandler(&huart2);
}

void USART3_IRQHandler(void)
{
    HAL_UART_IRQHandler(&huart3);
}

void UART4_IRQHandler(void)
{
    HAL_UART_IRQHandler(&huart4);
}

void DMA1_Stream4_IRQHandler(void)
{
    HAL_DMA_IRQHandler(&hdma_uart4_rx);
}

void TIM1_UP_TIM10_IRQHandler(void)
{
    HAL_TIM_IRQHandler(&htim1);
}

void TIM1_CC_IRQHandler(void)
{
    HAL_TIM_IRQHandler(&htim1);
}

void EXTI9_5_IRQHandler(void)
{
}

void EXTI15_10_IRQHandler(void)
{
}

void I2C1_EV_IRQHandler(void)
{
}

void I2C1_ER_IRQHandler(void)
{
}
