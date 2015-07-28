**Highly experimental and in development**

# cnt
Tool to build container


# cnt configuration file

cnt global conf is a `config.yml` file located in :
* windows >  $HOME + "/AppData/Local/Cnt";
* darwin > $HOME + "/Library/Cnt";
* linux > $HOME + "/.config/cnt";

with content :
```
push:
  type: maven
  url: https://localhost/nexus
  username: admin
  password: admin 

```