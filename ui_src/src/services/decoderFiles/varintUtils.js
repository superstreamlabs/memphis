import JSBI from "jsbi";

const BIGINT_1 = JSBI.BigInt(1);
const BIGINT_2 = JSBI.BigInt(2);

export function interpretAsSignedType(n) {
  // see https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/wire_format_lite.h#L857-L876
  // however, this is a simpler equivalent formula
  const isEven = JSBI.equal(JSBI.bitwiseAnd(n, JSBI.BigInt(1)), JSBI.BigInt(0));
  if (isEven) {
    return JSBI.divide(n, BIGINT_2);
  } else {
    return JSBI.multiply(
      JSBI.BigInt(-1),
      JSBI.divide(JSBI.add(n, BIGINT_1), BIGINT_2)
    );
  }
}

export function decodeVarint(buffer, offset) {
  let res = JSBI.BigInt(0);
  let shift = 0;
  let byte = 0;

  do {
    if (offset >= buffer.length) {
      throw new RangeError("Index out of bound decoding varint");
    }

    byte = buffer[offset++];

    const multiplier = JSBI.exponentiate(BIGINT_2, JSBI.BigInt(shift));
    const thisByteValue = JSBI.multiply(JSBI.BigInt(byte & 0x7f), multiplier);
    shift += 7;
    res = JSBI.add(res, thisByteValue);
  } while (byte >= 0x80);

  return {
    value: res,
    length: shift / 7
  };
}
