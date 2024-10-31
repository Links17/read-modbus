#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <modbus.h>
#include <errno.h>
#include <unistd.h>  // 引入用于 sleep 的库

#define START_ADDRESS 0x8007  // 从机保持寄存器起始地址
#define BATCH_SIZE 100        // 每次读取的寄存器数量上限

int main() {
    modbus_t *ctx;
    uint16_t reg_count = 0;  // 要读取的寄存器数量
    uint16_t *registers = NULL;  // 保存读取的数据
    FILE *file;
    int rc;
    int total_read = 0;

    // 1. 创建 Modbus RTU 上下文，串口路径和配置
    ctx = modbus_new_rtu("/dev/ttyUSB0", 115200, 'N', 8, 1);
    if (ctx == NULL) {
        fprintf(stderr, "无法创建 Modbus RTU 连接: %s\n", modbus_strerror(errno));
        return -1;
    }

    // 2. 设置 Modbus 从机地址（假设从机地址是 1）
    modbus_set_slave(ctx, 1);

    // 3. 连接到从机
    if (modbus_connect(ctx) == -1) {
        fprintf(stderr, "连接失败: %s\n", modbus_strerror(errno));
        modbus_free(ctx);
        return -1;
    }

    // 4. 读取寄存器 0x8002
    uint16_t dummy_register;
    rc = modbus_read_registers(ctx, 0x8002, 1, &dummy_register);  // 读取寄存器 0x8002
    if (rc == -1) {
        fprintf(stderr, "读取寄存器 0x8002 失败: %s\n", modbus_strerror(errno));
        modbus_close(ctx);
        modbus_free(ctx);
        return -1;
    }
    printf("读取寄存器 0x8002 成功，值为: %d\n", dummy_register);

    // 等待3秒
    sleep(3);

    // 5. 读取保持寄存器 0x8006，获取要读取的寄存器数量
    uint16_t count_register;
    rc = modbus_read_registers(ctx, 0x8006, 1, &count_register);  // 读取1个寄存器
    if (rc == -1) {
        fprintf(stderr, "读取保持寄存器失败: %s\n", modbus_strerror(errno));
        modbus_close(ctx);
        modbus_free(ctx);
        return -1;
    }
    reg_count = count_register;  // 要读取的寄存器数量

    printf("要读取的寄存器数量: %d\n", reg_count);

    // 以下代码保持不变
    // 6. 分配内存以保存所有寄存器数据
    registers = (uint16_t *)malloc(reg_count * sizeof(uint16_t));
    if (registers == NULL) {
        fprintf(stderr, "内存分配失败\n");
        modbus_close(ctx);
        modbus_free(ctx);
        return -1;
    }

    // 7. 分批读取保持寄存器
    int start_address = START_ADDRESS;
    int remaining_regs = reg_count;
    while (remaining_regs > 0) {
        int read_size = (remaining_regs > BATCH_SIZE) ? BATCH_SIZE : remaining_regs;

        rc = modbus_read_registers(ctx, start_address, read_size, &registers[total_read]);
        if (rc == -1) {
            fprintf(stderr, "读取保持寄存器失败: %s\n", modbus_strerror(errno));
            free(registers);
            modbus_close(ctx);
            modbus_free(ctx);
            return -1;
        }

        total_read += read_size;
        remaining_regs -= read_size;
        start_address += read_size;

        printf("已读取 %d 个寄存器\n", total_read);
    }

    // 8. 打开文件并写入字符
    file = fopen("modbus_data.txt", "w");
    if (file == NULL) {
        fprintf(stderr, "无法创建文件\n");
        free(registers);
        modbus_close(ctx);
        modbus_free(ctx);
        return -1;
    }

    // 9. 逐个寄存器解析，翻转字节序，然后将高字节和低字节转换为字符
    for (int i = 0; i < reg_count; i++) {
        uint16_t reg_value = registers[i];
        uint16_t flipped_value = (reg_value << 8) | (reg_value >> 8);
        char high_byte = (char)((flipped_value >> 8) & 0xFF);
        char low_byte = (char)(flipped_value & 0xFF);

        fprintf(file, "%c%c", high_byte, low_byte);
    }

    printf("所有字符数据已写入 modbus_data.txt 文件\n");

    // 10. 清理资源
    fclose(file);
    free(registers);
    modbus_close(ctx);
    modbus_free(ctx);

    return 0;
}