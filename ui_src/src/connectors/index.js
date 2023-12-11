import S3LogoIcon from './assets/s3LogoIcon.svg';
import KafkaIcon from './assets/kafkaIcon.svg';
import KinesisIcon from './assets/awsKinesis.svg';

import { kafka } from './kafka';
import { kinesis } from './kinesis';

export const connectorTypes = [
    { name: 'Kafka', icon: KafkaIcon, comment: 'Supported version: 0.9 and above', inputs: kafka },
    // { name: 'kinesis', icon: KinesisIcon, inputs: kinesis },
    { name: 'S3', icon: S3LogoIcon, disabled: true, soon: true }
];
