# Change Log

All notable changes to this project will be documented in this file.
See [Conventional Commits](https://conventionalcommits.org) for commit guidelines.



# [2.23.8](https://github.com/go-courier/sqlx/compare/v2.23.7...v2.23.8)


# [2.23.7](https://github.com/go-courier/sqlx/compare/v2.23.6...v2.23.7)

### Bug Fixes

* **fix(builder):** EachStructField fix ([1b71e9b](https://github.com/go-courier/sqlx/commit/1b71e9bdc7c9f37643b6322f357e635d907c3066))



# [2.23.6](https://github.com/go-courier/sqlx/compare/v2.23.5...v2.23.6)

### Bug Fixes

* **fix(builder):** loc pollution fix ([ad35502](https://github.com/go-courier/sqlx/commit/ad355029c3a8816bdee796c5c382d0efaa69c0f5))



# [2.23.5](https://github.com/go-courier/sqlx/compare/v2.23.4...v2.23.5)

### Bug Fixes

* **fix(builder):** reflectx should use x/reflect ([7d69bea](https://github.com/go-courier/sqlx/commit/7d69bea5f35b39862a0cc1fde4d980ee7476d148))



# [2.23.4](https://github.com/go-courier/sqlx/compare/v2.23.3...v2.23.4)

### Bug Fixes

* **fix(builder):** TableFiels should store loc of model value ([41843c2](https://github.com/go-courier/sqlx/commit/41843c2b8d86f4354355efc1ecc68e7e93f83bb8))



# [2.23.3](https://github.com/go-courier/sqlx/compare/v2.23.2...v2.23.3)

### Bug Fixes

* **fix** revert expr from context ([8c9f485](https://github.com/go-courier/sqlx/commit/8c9f48583a347156d1d7d5b5f713f3cceca4a229))



# [2.23.2](https://github.com/go-courier/sqlx/compare/v2.23.1...v2.23.2)



# [2.23.1](https://github.com/go-courier/sqlx/compare/v2.23.0...v2.23.1)

### Bug Fixes

* **fix(connectors/postgresql):** should have blank before index def ([b2601de](https://github.com/go-courier/sqlx/commit/b2601de9620b73151dd103acb38107ab1b1678cf))



# [2.23.0](https://github.com/go-courier/sqlx/compare/v2.22.0...v2.23.0)

### Features

* **feat** custom index def ([e5a928c](https://github.com/go-courier/sqlx/commit/e5a928c666d0234ecc5f1868afc7885e6477c94e))



# [2.22.0](https://github.com/go-courier/sqlx/compare/v2.21.6...v2.22.0)

### Features

* **feat** bump to golang 1.17 and deps updates ([17fcb58](https://github.com/go-courier/sqlx/commit/17fcb5811c03ca3ab2713030581586c5669115f6))



# [2.21.4](https://github.com/go-courier/sqlx/compare/v2.21.3...v2.21.4)

### Bug Fixes

* **fix** should ignore deprecated field value ([9d393eb](https://github.com/go-courier/sqlx/commit/9d393eb782d773ae2c7bce4abdfde0401770a4c1))



# [2.21.3](https://github.com/go-courier/sqlx/compare/v2.21.2...v2.21.3)

### Bug Fixes

* **fix** should ignore deprecated field value ([acdc920](https://github.com/go-courier/sqlx/commit/acdc9205aa14e11b0ef06f24226232d8d9fc2a52))



# [2.21.2](https://github.com/go-courier/sqlx/compare/v2.21.1...v2.21.2)

### Bug Fixes

* **fix** should UnwrapAll before db err check. ([6503ee0](https://github.com/go-courier/sqlx/commit/6503ee04c06eb296fcbc9c2e8b0abe5ad8ea263a))



# [2.21.1](https://github.com/go-courier/sqlx/compare/v2.21.0...v2.21.1)

### Bug Fixes

* **fix** driver connect issue when ctx pass may be cancel ([da0dbe4](https://github.com/go-courier/sqlx/commit/da0dbe4cbbdea220082fa6a6fa64a7b04edf7c22))



# [2.21.0](https://github.com/go-courier/sqlx/compare/v2.20.5...v2.21.0)

### Features

* **feat** drop logrus ([b4a9e7b](https://github.com/go-courier/sqlx/commit/b4a9e7ba17de52967d6d29064fa97cfee05a3383))



# [2.20.5](https://github.com/go-courier/sqlx/compare/v2.20.4...v2.20.5)

### Bug Fixes

* **fix(pg):** avoid gen invalid cmd ([42e911f](https://github.com/go-courier/sqlx/commit/42e911fcdfcba0f5501fa5878e0b2e9654f23fcb))



# [2.20.4](https://github.com/go-courier/sqlx/compare/v2.20.3...v2.20.4)

### Bug Fixes

* **fix(migration):** create table should dry run ([93c0b30](https://github.com/go-courier/sqlx/commit/93c0b304938fa4ef8971412891d77964f6c80f5a))



# [2.20.3](https://github.com/go-courier/sqlx/compare/v2.20.2...v2.20.3)

### Bug Fixes

* **fix(pg):** default number should be with quote and dataType ([4c0575b](https://github.com/go-courier/sqlx/commit/4c0575bf3cc9a1b03ee4715dd0fc9ffabebeb4ed))



# [2.20.2](https://github.com/go-courier/sqlx/compare/v2.20.1...v2.20.2)

### Bug Fixes

* **fix** pg comma fix ([9b52cbe](https://github.com/go-courier/sqlx/commit/9b52cbebed06597a72fe5feff207aacc1b803583))



# [2.20.1](https://github.com/go-courier/sqlx/compare/v2.20.0...v2.20.1)

### Bug Fixes

* **fix(migration):** log prev default value ([1cce629](https://github.com/go-courier/sqlx/commit/1cce629042e1877fcdb9c3f8304b00af547444cc))



# [2.20.0](https://github.com/go-courier/sqlx/compare/v2.19.1...v2.20.0)

### Features

* **feat** migration enhancement ([1fa4a9e](https://github.com/go-courier/sqlx/commit/1fa4a9e92cf79f359f3820763eeaad36b8666ea1))



# [2.19.1](https://github.com/go-courier/sqlx/compare/v2.19.0...v2.19.1)

### Bug Fixes

* **fix** enhance GetColumnName ([d7074af](https://github.com/go-courier/sqlx/commit/d7074af444c88c61f547a37636d294b18b94c0ee))



# [2.19.0](https://github.com/go-courier/sqlx/compare/v2.18.2...v2.19.0)

### Features

* **feat** alias tag for table join rename ([fc2ed04](https://github.com/go-courier/sqlx/commit/fc2ed0417fadb395d937b684bf31b1425051b924))



# [2.18.2](https://github.com/go-courier/sqlx/compare/v2.18.1...v2.18.2)

### Bug Fixes

* **fix(builder):** fieldValue fix ([ab91aec](https://github.com/go-courier/sqlx/commit/ab91aecbb684be35ec597eae670759039ac70c6b))



# [2.18.1](https://github.com/go-courier/sqlx/compare/v2.18.0...v2.18.1)

### Bug Fixes

* **fix(builder):** only struct could be visited ([ef21879](https://github.com/go-courier/sqlx/commit/ef218797c8245462060a6fd45b4bf6cc84e81cd2))



# [2.18.0](https://github.com/go-courier/sqlx/compare/v2.17.4...v2.18.0)

### Features

* **feat** support scan joined values ([f160fe4](https://github.com/go-courier/sqlx/commit/f160fe4750d488e50566c475c88069e2c1de652b))
