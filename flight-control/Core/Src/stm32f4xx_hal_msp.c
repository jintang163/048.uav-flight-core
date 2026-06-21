#include "main.h"

void HAL_MspInit(void)
{
    __HAL_RCC_SYSCFG_CLK_ENABLE();
    __HAL_RCC_PWR_CLK_ENABLE();

    HAL_NVIC_SetPriority(PendSV_IRQn, 15, 0);
}

void HAL_I2C_MspInit(I2C_HandleTypeDef* hi2c)
{
    GPIO_InitTypeDef GPIO_InitStruct = {0};
    if (hi2c->Instance == I2C1) {
        __HAL_RCC_GPIOB_CLK_ENABLE();

        GPIO_InitStruct.Pin = GPIO_PIN_6 | GPIO_PIN_7;
        GPIO_InitStruct.Mode = GPIO_MODE_AF_OD;
        GPIO_InitStruct.Pull = GPIO_PULLUP;
        GPIO_InitStruct.Speed = GPIO_SPEED_FREQ_VERY_HIGH;
        GPIO_InitStruct.Alternate = GPIO_AF4_I2C1;
        HAL_GPIO_Init(GPIOB, &GPIO_InitStruct);

        __HAL_RCC_I2C1_CLK_ENABLE();

        HAL_NVIC_SetPriority(I2C1_EV_IRQn, 5, 0);
        HAL_NVIC_EnableIRQ(I2C1_EV_IRQn);
        HAL_NVIC_SetPriority(I2C1_ER_IRQn, 5, 0);
        HAL_NVIC_EnableIRQ(I2C1_ER_IRQn);
    }
}

void HAL_I2C_MspDeInit(I2C_HandleTypeDef* hi2c)
{
    if (hi2c->Instance == I2C1) {
        __HAL_RCC_I2C1_CLK_DISABLE();
        HAL_GPIO_DeInit(GPIOB, GPIO_PIN_6 | GPIO_PIN_7);
        HAL_NVIC_DisableIRQ(I2C1_EV_IRQn);
        HAL_NVIC_DisableIRQ(I2C1_ER_IRQn);
    }
}

void HAL_TIM_OC_MspInit(TIM_HandleTypeDef* htim_oc)
{
    GPIO_InitTypeDef GPIO_InitStruct = {0};
    if (htim_oc->Instance == TIM1) {
        __HAL_RCC_TIM1_CLK_ENABLE();
        __HAL_RCC_GPIOA_CLK_ENABLE();

        GPIO_InitStruct.Pin = GPIO_PIN_8 | GPIO_PIN_9 | GPIO_PIN_10 | GPIO_PIN_11;
        GPIO_InitStruct.Mode = GPIO_MODE_AF_PP;
        GPIO_InitStruct.Pull = GPIO_NOPULL;
        GPIO_InitStruct.Speed = GPIO_SPEED_FREQ_LOW;
        GPIO_InitStruct.Alternate = GPIO_AF1_TIM1;
        HAL_GPIO_Init(GPIOA, &GPIO_InitStruct);

        HAL_NVIC_SetPriority(TIM1_UP_TIM10_IRQn, 6, 0);
        HAL_NVIC_EnableIRQ(TIM1_UP_TIM10_IRQn);
        HAL_NVIC_SetPriority(TIM1_CC_IRQn, 6, 0);
        HAL_NVIC_EnableIRQ(TIM1_CC_IRQn);
    }
}

void HAL_TIM_OC_MspDeInit(TIM_HandleTypeDef* htim_oc)
{
    if (htim_oc->Instance == TIM1) {
        __HAL_RCC_TIM1_CLK_DISABLE();
        HAL_GPIO_DeInit(GPIOA, GPIO_PIN_8 | GPIO_PIN_9 | GPIO_PIN_10 | GPIO_PIN_11);
        HAL_NVIC_DisableIRQ(TIM1_UP_TIM10_IRQn);
        HAL_NVIC_DisableIRQ(TIM1_CC_IRQn);
    }
}

void HAL_UART_MspInit(UART_HandleTypeDef* huart)
{
    GPIO_InitTypeDef GPIO_InitStruct = {0};
    if (huart->Instance == USART1) {
        __HAL_RCC_USART1_CLK_ENABLE();
        __HAL_RCC_GPIOA_CLK_ENABLE();

        GPIO_InitStruct.Pin = GPIO_PIN_9 | GPIO_PIN_10;
        GPIO_InitStruct.Mode = GPIO_MODE_AF_PP;
        GPIO_InitStruct.Pull = GPIO_NOPULL;
        GPIO_InitStruct.Speed = GPIO_SPEED_FREQ_VERY_HIGH;
        GPIO_InitStruct.Alternate = GPIO_AF7_USART1;
        HAL_GPIO_Init(GPIOA, &GPIO_InitStruct);

        __HAL_RCC_DMA2_CLK_ENABLE();

        hdma_usart1_rx.Instance = DMA2_Stream2;
        hdma_usart1_rx.Init.Channel = DMA_CHANNEL_4;
        hdma_usart1_rx.Init.Direction = DMA_PERIPH_TO_MEMORY;
        hdma_usart1_rx.Init.PeriphInc = DMA_PINC_DISABLE;
        hdma_usart1_rx.Init.MemInc = DMA_MINC_ENABLE;
        hdma_usart1_rx.Init.PeriphDataAlignment = DMA_PDATAALIGN_BYTE;
        hdma_usart1_rx.Init.MemDataAlignment = DMA_MDATAALIGN_BYTE;
        hdma_usart1_rx.Init.Mode = DMA_CIRCULAR;
        hdma_usart1_rx.Init.Priority = DMA_PRIORITY_HIGH;
        hdma_usart1_rx.Init.FIFOMode = DMA_FIFOMODE_DISABLE;
        HAL_DMA_Init(&hdma_usart1_rx);

        __HAL_LINKDMA(huart, hdmarx, hdma_usart1_rx);

        HAL_NVIC_SetPriority(DMA2_Stream2_IRQn, 4, 0);
        HAL_NVIC_EnableIRQ(DMA2_Stream2_IRQn);
        HAL_NVIC_SetPriority(USART1_IRQn, 5, 0);
        HAL_NVIC_EnableIRQ(USART1_IRQn);
    } else if (huart->Instance == USART2) {
        __HAL_RCC_USART2_CLK_ENABLE();
        __HAL_RCC_GPIOA_CLK_ENABLE();

        GPIO_InitStruct.Pin = GPIO_PIN_2 | GPIO_PIN_3;
        GPIO_InitStruct.Mode = GPIO_MODE_AF_PP;
        GPIO_InitStruct.Pull = GPIO_NOPULL;
        GPIO_InitStruct.Speed = GPIO_SPEED_FREQ_VERY_HIGH;
        GPIO_InitStruct.Alternate = GPIO_AF7_USART2;
        HAL_GPIO_Init(GPIOA, &GPIO_InitStruct);

        __HAL_RCC_DMA1_CLK_ENABLE();

        hdma_usart2_rx.Instance = DMA1_Stream5;
        hdma_usart2_rx.Init.Channel = DMA_CHANNEL_4;
        hdma_usart2_rx.Init.Direction = DMA_PERIPH_TO_MEMORY;
        hdma_usart2_rx.Init.PeriphInc = DMA_PINC_DISABLE;
        hdma_usart2_rx.Init.MemInc = DMA_MINC_ENABLE;
        hdma_usart2_rx.Init.PeriphDataAlignment = DMA_PDATAALIGN_BYTE;
        hdma_usart2_rx.Init.MemDataAlignment = DMA_MDATAALIGN_BYTE;
        hdma_usart2_rx.Init.Mode = DMA_CIRCULAR;
        hdma_usart2_rx.Init.Priority = DMA_PRIORITY_HIGH;
        hdma_usart2_rx.Init.FIFOMode = DMA_FIFOMODE_DISABLE;
        HAL_DMA_Init(&hdma_usart2_rx);

        __HAL_LINKDMA(huart, hdmarx, hdma_usart2_rx);

        HAL_NVIC_SetPriority(DMA1_Stream5_IRQn, 4, 0);
        HAL_NVIC_EnableIRQ(DMA1_Stream5_IRQn);
        HAL_NVIC_SetPriority(USART2_IRQn, 5, 0);
        HAL_NVIC_EnableIRQ(USART2_IRQn);
    } else if (huart->Instance == USART3) {
        __HAL_RCC_USART3_CLK_ENABLE();
        __HAL_RCC_GPIOB_CLK_ENABLE();

        GPIO_InitStruct.Pin = GPIO_PIN_10 | GPIO_PIN_11;
        GPIO_InitStruct.Mode = GPIO_MODE_AF_PP;
        GPIO_InitStruct.Pull = GPIO_NOPULL;
        GPIO_InitStruct.Speed = GPIO_SPEED_FREQ_VERY_HIGH;
        GPIO_InitStruct.Alternate = GPIO_AF7_USART3;
        HAL_GPIO_Init(GPIOB, &GPIO_InitStruct);

        hdma_usart3_rx.Instance = DMA1_Stream6;
        hdma_usart3_rx.Init.Channel = DMA_CHANNEL_5;
        hdma_usart3_rx.Init.Direction = DMA_PERIPH_TO_MEMORY;
        hdma_usart3_rx.Init.PeriphInc = DMA_PINC_DISABLE;
        hdma_usart3_rx.Init.MemInc = DMA_MINC_ENABLE;
        hdma_usart3_rx.Init.PeriphDataAlignment = DMA_PDATAALIGN_BYTE;
        hdma_usart3_rx.Init.MemDataAlignment = DMA_MDATAALIGN_BYTE;
        hdma_usart3_rx.Init.Mode = DMA_CIRCULAR;
        hdma_usart3_rx.Init.Priority = DMA_PRIORITY_HIGH;
        hdma_usart3_rx.Init.FIFOMode = DMA_FIFOMODE_DISABLE;
        HAL_DMA_Init(&hdma_usart3_rx);

        __HAL_LINKDMA(huart, hdmarx, hdma_usart3_rx);

        HAL_NVIC_SetPriority(DMA1_Stream6_IRQn, 4, 0);
        HAL_NVIC_EnableIRQ(DMA1_Stream6_IRQn);
        HAL_NVIC_SetPriority(USART3_IRQn, 5, 0);
        HAL_NVIC_EnableIRQ(USART3_IRQn);
    } else if (huart->Instance == UART4) {
        __HAL_RCC_UART4_CLK_ENABLE();
        __HAL_RCC_GPIOC_CLK_ENABLE();

        GPIO_InitStruct.Pin = GPIO_PIN_10 | GPIO_PIN_11;
        GPIO_InitStruct.Mode = GPIO_MODE_AF_PP;
        GPIO_InitStruct.Pull = GPIO_NOPULL;
        GPIO_InitStruct.Speed = GPIO_SPEED_FREQ_VERY_HIGH;
        GPIO_InitStruct.Alternate = GPIO_AF8_UART4;
        HAL_GPIO_Init(GPIOC, &GPIO_InitStruct);

        hdma_uart4_rx.Instance = DMA1_Stream4;
        hdma_uart4_rx.Init.Channel = DMA_CHANNEL_4;
        hdma_uart4_rx.Init.Direction = DMA_PERIPH_TO_MEMORY;
        hdma_uart4_rx.Init.PeriphInc = DMA_PINC_DISABLE;
        hdma_uart4_rx.Init.MemInc = DMA_MINC_ENABLE;
        hdma_uart4_rx.Init.PeriphDataAlignment = DMA_PDATAALIGN_BYTE;
        hdma_uart4_rx.Init.MemDataAlignment = DMA_MDATAALIGN_BYTE;
        hdma_uart4_rx.Init.Mode = DMA_CIRCULAR;
        hdma_uart4_rx.Init.Priority = DMA_PRIORITY_HIGH;
        hdma_uart4_rx.Init.FIFOMode = DMA_FIFOMODE_DISABLE;
        HAL_DMA_Init(&hdma_uart4_rx);

        __HAL_LINKDMA(huart, hdmarx, hdma_uart4_rx);

        HAL_NVIC_SetPriority(DMA1_Stream4_IRQn, 4, 0);
        HAL_NVIC_EnableIRQ(DMA1_Stream4_IRQn);
        HAL_NVIC_SetPriority(UART4_IRQn, 5, 0);
        HAL_NVIC_EnableIRQ(UART4_IRQn);
    }
}

void HAL_UART_MspDeInit(UART_HandleTypeDef* huart)
{
    if (huart->Instance == USART1) {
        __HAL_RCC_USART1_CLK_DISABLE();
        HAL_GPIO_DeInit(GPIOA, GPIO_PIN_9 | GPIO_PIN_10);
        HAL_DMA_DeInit(huart->hdmarx);
        HAL_NVIC_DisableIRQ(DMA2_Stream2_IRQn);
        HAL_NVIC_DisableIRQ(USART1_IRQn);
    } else if (huart->Instance == USART2) {
        __HAL_RCC_USART2_CLK_DISABLE();
        HAL_GPIO_DeInit(GPIOA, GPIO_PIN_2 | GPIO_PIN_3);
        HAL_DMA_DeInit(huart->hdmarx);
        HAL_NVIC_DisableIRQ(DMA1_Stream5_IRQn);
        HAL_NVIC_DisableIRQ(USART2_IRQn);
    } else if (huart->Instance == USART3) {
        __HAL_RCC_USART3_CLK_DISABLE();
        HAL_GPIO_DeInit(GPIOB, GPIO_PIN_10 | GPIO_PIN_11);
        HAL_DMA_DeInit(huart->hdmarx);
        HAL_NVIC_DisableIRQ(DMA1_Stream6_IRQn);
        HAL_NVIC_DisableIRQ(USART3_IRQn);
    } else if (huart->Instance == UART4) {
        __HAL_RCC_UART4_CLK_DISABLE();
        HAL_GPIO_DeInit(GPIOC, GPIO_PIN_10 | GPIO_PIN_11);
        HAL_DMA_DeInit(huart->hdmarx);
        HAL_NVIC_DisableIRQ(DMA1_Stream4_IRQn);
        HAL_NVIC_DisableIRQ(UART4_IRQn);
    }
}
