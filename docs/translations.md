## Translations guidelines

To update translations proceed with the below workflow: 
```
goi18n extract
```
create file translate.*.toml 
Note: * is used instead of a particular language code. for example create `translate.ru.toml` file if you want to modify Russian translations
```
goi18n merge active.*.toml translate.*.toml
```
after you translate all the message 
```
goi18n merge active.*.toml translate.*.toml
```