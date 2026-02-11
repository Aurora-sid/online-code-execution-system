FROM mcr.microsoft.com/dotnet/sdk:8.0

WORKDIR /app

# 创建一个空的控制台项目作为模板，利用缓存层的nuget包
RUN dotnet new console -o .
RUN dotnet build

# 清理
RUN rm -rf obj bin Program.cs

# 保持容器运行
CMD ["sleep", "infinity"]
