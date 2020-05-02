# Locking-Center

Locking-Center is a mutex point to synchronize access between different services. You can limit the 
execution between services and create queueing for the operation.

#### What is it for?
Let's assume that you are keeping a text file in a common file location. One service wants to reach it and
reads some data and appends at the end of it and other service wants to find a part of the file and removes it. If you
try to reach this file from different services at the same time, it creates race condition and one of those services
will fail. You should somehow limit the access of services to this file.

One way to success on is to create a file manipulation service so other services will access to the file manipulation
service and this service will manage the synchronization. However, this will not be a scalable solution. If you 
want to scale, you should create a message queue and multi file manipulation services will handle in sync. But this
time you will not have real time reflection for changes because you will not have exact knowledge for it. So
you should create another queue to notify the services which are waiting for the result of the file changes. This
scenario works if you do not need to know the change result immediately.

One another way to success on is to caching. For example Redis is a good choice to keep in memory access to the data
with the help of locking feature of the service. So periodically, you can write the changes to file on the disk from
Redis cache. However, file can be too big to keep in the memory all the time and creates memory efficiency problem. 

On this point, why Locking-Center is needed. Locking-Center provides very useful primitive in many environments where
different processes must operate with shared resources in a mutually exclusive way. You can Lock the partial of the code
with Locking-Center and that shared resource will not be accessible until the Lock owner releases the Lock. So, it
provides, memory efficiency, able to know about the result and change reflection immediately and no need to create
complicated queue architecture in the whole system.

#### Setup Preparation

- Download the latest release of Locking-Center or compile it using the `create_release.sh` shell script file located
under the `-build-` folder.

##### Setting Up

- Copy `locking-center` executable to `/usr/local/bin` folder on the system.
- Give execution permission to the file `sudo chmod +x /usr/local/bin/locking-center`
- Create an empty file in your user path, copy-paste the following and save the file
```shell script
#!/bin/sh

export BIND_ADDRESS="localhost:22119" # This is optional, if it is not defined it will be `:22119`
/usr/local/bin/locking-center
```
- Give execution permission to the file `sudo chmod +x [Saved File Location]`
- Execute the saved file.
---
##### Mutex Usage

you can use the clients for the service
- [https://github.com/freakmaxi/locking-center-client-csharp][C#]
- [https://github.com/freakmaxi/locking-center-client-go][Go]

or

you can follow the instruction.

Locking-Center is working using TCP Connection. When you make the first TCP connection, you should send the details
package.

##### Locking

Package is consist of byte(key string size)/string(key string)/byte(action type) format. Let's create a package.

Ex: key to lock is `locking-me`. **Lock key should not be more than 128 characters or empty/null string.**

 `locking-me` is 10 chars. First byte of the package is the length of the key string. so first byte is 10.
 
 String is `locking-me` in UTF-8 encoding.
 
 Action Type is the requested action. 
 
 - 1 = locking
 - 2 = unlocking
 - 3 = reset lock
 
 We want to lock, so the action type byte will be `1`
 
 at the end, you will create a byte array like this.

Package Byte Array: `[10, 108, 111, 99, 107, 105, 110, 103, 45, 109, 101, 1]`

When you make the request, you can receive 2 type of answers `-` or `+`

- `-` means operation is unsuccessful due to internal error like wrong key format or unlimited resource and try again.
- '+' means operation is successful and go on for the operation. 

if you get `-` you can check the key for the wrong format, if not, try again until you get `+`.

When you make the request, you may not get the answer immediately. It means, that key has been already locked and
waiting the lock to be released. Because of that, you should hang there until you get the answer. According to the
answer you should continue the operation and close the client connection.

**IMPORTANT: Because of this wait, your TCP client should not have timeout on the connection.** 

##### Unlocking

After you complete the operation on the shared resource, you should make another request to unlock the resource.

The same logic working in here. for the same key `locking-me` you should create a message package for unlocking. this
time it will be;

Package Byte Array: `[10, 108, 111, 99, 107, 105, 110, 103, 45, 109, 101, 2]`

When you make the request, you can receive 2 type of answers `-` or `+`

- `-` means operation is unsuccessful due to internal error like wrong key format or unlimited resource and try again.
- '+' means operation is successful and go on for the operation. 

if you get `-` you can check the key for the wrong format, if not, try again until you get `+`.

##### Lock Resetting

You may have a failure on your service and while it has lock, it can crash. Service crash will not release the lock 
automatically, you should send a reset request to the Locking-Center to drop all the locks for the key.

For this, you need to create a package as before with the different action type byte. Let's create a reset request
message package byte array for the same key `locking-me`.

Package Byte Array: `[10, 108, 111, 99, 107, 105, 110, 103, 45, 109, 101, 3]`

Last byte of the array is this time `3` because resetting action type is `3`.

**You can create your own client for the programming language that you are using and share with me, I'll put it in here on
client section.**

[Go]: https://github.com/freakmaxi/locking-center-client-go

[C#]: https://github.com/freakmaxi/locking-center-client-csharp