import JSBI from 'jsbi';
import { bufferLeToBeHex } from './hexUtils';

export function decodeFixed32(value) {
    const floatValue = value.readFloatLE(0);
    const intValue = value.readInt32LE(0);
    const uintValue = value.readUInt32LE(0);

    const result = [];

    result.push({ type: 'Int', value: intValue });

    if (intValue !== uintValue) {
        result.push({ type: 'Unsigned Int', value: uintValue });
    }

    result.push({ type: 'Float', value: floatValue });

    return result;
}

export function decodeFixed64(value) {
    const floatValue = value.readDoubleLE(0);
    const uintValue = JSBI.BigInt('0x' + bufferLeToBeHex(value));
    const intValue = twoComplements(uintValue);

    const result = [];

    result.push({ type: 'Int', value: intValue.toString() });

    if (intValue !== uintValue) {
        result.push({ type: 'Unsigned Int', value: uintValue.toString() });
    }

    result.push({ type: 'Double', value: floatValue });

    return result;
}

export function decodeVarintParts(value) {
    const result = [];
    const intVal = JSBI.BigInt(value);
    result.push({ type: 'Int', value: intVal.toString() });

    // const signedIntVal = interpretAsSignedType(intVal);
    // if (signedIntVal !== intVal) {
    //   result.push({ type: "Signed Int", value: signedIntVal.toString() });
    // }
    return result;
}

const maxLong = JSBI.BigInt('0x7fffffffffffffff');
const longForComplement = JSBI.BigInt('0x10000000000000000');

function twoComplements(uintValue) {
    if (JSBI.greaterThan(uintValue, maxLong)) {
        return JSBI.subtract(uintValue, longForComplement);
    } else {
        return uintValue;
    }
}
