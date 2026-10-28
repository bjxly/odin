/*
 * SmartJ 企业级多租户基础框架
 * Copyright (C) 2024-2026 SmartJ Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package com.smartj.odin;

import org.apache.ibatis.annotations.Mapper;
import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * SmartG-Odin 启动类。
 *
 * <p>复用 SmartJ 运行时（登录鉴权、系统管理、日志、缓存），并装载 odin 业务模块。
 * 组件扫描与 Mapper 扫描均覆盖 {@code com.smartj}，同时命中 smartj 与 odin（com.smartj.odin）包。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@SpringBootApplication(scanBasePackages = "com.smartj")
@MapperScan(basePackages = "com.smartj", annotationClass = Mapper.class)
public class SmartjOdinApplication {

    /**
     * 应用入口。
     *
     * @param args 启动参数
     */
    public static void main(String[] args) {
        System.out.println("当前程序路径：" + System.getProperty("user.dir"));
        SpringApplication.run(SmartjOdinApplication.class, args);
    }
}
