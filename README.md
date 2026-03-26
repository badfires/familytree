**FamilyTree APP**

------

This is a family tree display project developed in Golang, which can be used to showcase and manage your family history and member relationships.

------

## Features

Visual display of members across multiple generations of a family
The relationship links among family members
Personal detailed information record
Optional login verification mechanism
Support the display of original parents and stepparents information during adoption

## Quick Start

download bin/familytree.exe  run is ok

The encrypted string in the database is "abcd"
The administrator password is "abcd@1234"

There is no need to download family.db, as it will be automatically created during the first run. This file is very important and needs to be saved properly



## Build

wails build -ldflags "-X family-tree/handler.AdminPassword=abcd@1234"



**Display**

![](E:\code\family-tree\image.png)