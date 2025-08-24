# Bilinovel Downloader

这是一个用于从 Bilinovel 下载和生成轻小说 EPUB 电子书的工具。
生成的 EPUB 文件完全符合 EPUB 标准，可以在 Calibre 检查中无错误通过。

## 使用示例

1. 下载整本 `https://www.bilinovel.com/novel/2388.html`

   ```bash
   bilinovel-downloader download -n 2388
   ```

2. 下载单卷 `https://www.bilinovel.com/novel/2388/vol_84522.html`

   ```bash
   bilinovel-downloader download -n 2388 -v 84522
   ```

3. 对自动生成的 epub 格式不满意可以自行修改后使用命令打包

   ```bash
   bilinovel-downloader pack -d <目录路径>
   ```
