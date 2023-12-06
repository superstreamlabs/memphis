import S3LogoIcon from './assets/s3LogoIcon.svg';
import KafkaIcon from './assets/kafkaIcon.svg';
import KinesisIcon from './assets/awsKinesis.svg';

import { kafka } from './kafka';
import { kinesis } from './kinesis';

export const connectorTypes = [
    { name: 'kafka', icon: KafkaIcon, comment: 'Supported version: v1.0.3', inputs: kafka },
    // { name: 'kinesis', icon: KinesisIcon, inputs: kinesis },
    { name: 's3', icon: S3LogoIcon, disabled: true }
];
