import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useEditorStore = defineStore('editor', () => {
    const currentLanguage = ref('python')

    // 代码片段
    const snippets = {
        cpp: '#include <iostream>\n\nint main() {\n    std::cout << "Hello Aurora Code from C++17!" << std::endl;\n    return 0;\n}',
        c: '#include <stdio.h>\n\nint main() {\n    printf("Hello Aurora Code from C (gcc8.3.0)!\\n");\n    return 0;\n}',
        java: 'public class Main {\n    public static void main(String[] args) {\n        System.out.println("Hello Aurora Code from Java 11!");\n    }\n}',
        python: 'print("Hello Aurora Code from Python 3.7.3!")',
        go: 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello Aurora Code from Go 1.19.5")\n}',
        javascript: 'console.log("Hello Aurora Code from JavaScript (Node.js)!");',
        rust: 'fn main() {\n    println!("Hello Aurora Code from Rust!");\n}',
        csharp: 'using System;\n\nclass Program {\n    static void Main() {\n        Console.WriteLine("Hello Aurora Code from C# (.NET 8)!");\n    }\n}',
        typescript: 'const msg: string = "Hello Aurora Code from TypeScript!";\nconsole.log(msg);'
    }

    // 存储每个语言的代码内容
    const codeMap = ref({})

    // 初始化代码内容（使用本地存储或默认片段）
    const loadCode = () => {
        // 这里可以扩展为从 localStorage 加载
    }

    // 更新代码
    const updateCode = (lang, newCode) => {
        codeMap.value[lang] = newCode
    }

    // 终端输出
    const terminalOutput = ref('')

    const clearOutput = () => {
        terminalOutput.value = ''
    }

    const appendOutput = (text) => {
        terminalOutput.value += text
    }

    return {
        currentLanguage,
        snippets,
        codeMap,
        loadCode,
        updateCode,
        terminalOutput,
        clearOutput,
        appendOutput
    }
})
