这是一个非常简陋的版本, 我不是研发出身, 我出于对家族的历史兴趣产生了这个想法

下载之后针对数据库有加密 密码 是abcd
编译的采用如下语句,
wails build -ldflags "-X family-tree/handler.AdminPassword=abcd@1234"
打开的时候管理密码是abcd@1234
