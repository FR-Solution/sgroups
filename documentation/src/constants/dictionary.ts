import { TDictionary } from '../customTypes/dictionary'

export const DICTIONARY: TDictionary = {
  syncOp: {
    short: 'Поле определяющее действие с данными из запроса.',
    full: '',
  },

  traffic: {
    short: 'Поле описывающий направление трафика.',
    full: '',
  },

  transport: {
    short: 'Протокол L3/L4 уровня модели OSI.',
    full: '',
  },

  log: {
    short: 'Включить/отключить логирование.',
    full: '',
  },

  trace: {
    short: 'Включить/отключить трассировку.',
    full: '',
  },

  ports: {
    short: 'Блок описывающий набор пар портов (src-dst).',
    full: '',
  },

  srcPorts: {
    short: 'Набор открытых портов отправителя.',
    full: '',
  },

  dstPorts: {
    short: 'Набор открытых портов получателя',
    full: '',
  },

  apiIcmp: {
    short: 'Структура, содержащая описание создаваемых правил типа ICMP.',
    full: '',
  },

  icmpV: {
    short: 'Версия IP для ICMP (IPv4 или IPv6).',
    full: '',
  },

  icmpTypes: {
    short: 'Список, определяющий допустимые типы ICMP запросов.',
    full: '',
  },

  sgroupSet: {
    short: 'Список, содержащий названия Security Group(s).',
    full: '',
  },

  sg: {
    short: 'Security Group, с которой устанавливаются правила взаимодействия.',
    full: '',
  },

  sgLocal: {
    short: 'Security Group относительно которой рассматриваются правила.',
    full: '',
  },

  description: {
    short: 'Формальное текстовое описание.',
    full: '',
  },

  rules: {
    short: 'Структура, содержащая описание создаваемых правил.',
    full: '',
  },

  nftRuleType: {
    short: 'Характеристика описывающая, что принимается трафик типа ip.',
    full: '',
  },

  nftCounter: {
    short: 'Счетчик количества байтов и пакетов.',
    full: '',
  },

  nftRuleVerdict: {
    short: 'Результат применения правила, определяющий действие, которое будет применено к пакету.',
    full: '',
  },

  terraformModule: {
    short: '',
    full: `Terraform module представляет собой высокоуровневую абстракцию над terraform resources, которая
        упрощает работу с ресурсами Terraform, скрывая сложность их непосредственного использования. Он предлагает
        простой и понятный интерфейс для взаимодействия, позволяя пользователям легко интегрироваться и управлять
        компонентами инфраструктуры без необходимости глубоко погружаться в детали каждого ресурса.`,
  },

  terraformResource: {
    short: '',
    full: `Terraform resource является ключевым элементом в Terraform, предназначенным для управления различными
        аспектами инфраструктуры через код. Он позволяет задавать, настраивать и управлять инфраструктурными
        компонентами без привязки к их конкретным типам, обеспечивая автоматизацию развертывания и поддержки
        инфраструктуры согласно подходу Infrastructure as Code (IaC).`,
  },

  cidrSet: {
    short: 'Список, содержащий подсети типа IP.',
    full: '',
  },

  fqdnSet: {
    short: 'Список, содержащий FQDN записи.',
    full: '',
  },

  fqdn: {
    short: 'Полное доменное имя (FQDN), для которого применяется данное правило.',
    full: '',
  },

  l7ProtocolList: {
    short: 'Список протоколов L7 уровня модели OSI.',
    full: '',
  },

  networks: {
    short: 'Массив/Список подсетей типа IP.',
    full: '',
  },

  nw: {
    short: 'Имя подсети',
    full: '',
  },

  networkNames: {
    short: 'Массив/Список имен подсетей',
    full: '',
  },

  networkObject: {
    short: 'Структура, содержащая описание сети',
    full: '',
  },

  cidr: {
    short: 'Подсеть типа IP.',
    full: '',
  },

  srcDstCidr: {
    short: 'CIDR, с которой устанавливаются правила взаимодействия.',
    full: '',
  },

  terraformItems: {
    short: 'Список ресурсов создаваемые terraform ресурсом.',
    full: '',
  },

  terraformRuleName: {
    short: 'Уникальное имя создаваемого ресурса.',
    full: '',
  },

  defaultAction: {
    short: 'Действие по умолчанию.',
    full: '',
  },

  action: {
    short: 'Действие для пакетов в сформированных правил в цепочке.',
    full: '',
  },

  priority: {
    short: 'Поле определяющее порядок применения правил в цепочке.',
    full: '',
  },

  priorityst: {
    short: 'Структура, содержащая описание порядка применения правил в цепочке.',
    full: '',
  },
}
