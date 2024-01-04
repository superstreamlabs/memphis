import S3LogoIcon from './assets/s3LogoIcon.svg';
import KafkaIcon from './assets/kafkaIcon.svg';
import RedisIcon from './assets/redisIcon.svg';
import MemphisIcon from './assets/memphisIcon.svg';

import { kafka } from './kafka';
import { redis } from './redis';
import { memphis } from './memphis';

export const connectorTypesSource = [
    { name: 'Kafka', icon: KafkaIcon, comment: 'Supported version: 0.9 and above', inputs: kafka },
    { name: 'S3', icon: S3LogoIcon, disabled: true, soon: true }
];

export const connectorTypesSink = [
    { name: 'Kafka', icon: KafkaIcon, comment: 'Supported version: 0.9 and above', inputs: kafka },
    { name: 'Redis', icon: RedisIcon, inputs: redis },
    { name: 'Memphis',icon: MemphisIcon, inputs: memphis },  
    { name: 'S3', icon: S3LogoIcon, disabled: true, soon: true }
];
