FAMILYTREE_DB_KEY='abcd'
set FAMILYTREE_ADMIN_PASSWORD=123456

正式编译
wails build -ldflags "-X family-tree/handler.AdminPassword=abcd@1234"