import { decodeVarint } from './varintUtils';

export class BufferReader {
    constructor(buffer) {
        this.buffer = buffer;
        this.offset = 0;
    }

    readVarInt() {
        const result = decodeVarint(this.buffer, this.offset);
        this.offset += result.length;

        return result.value;
    }

    readBuffer(length) {
        this.checkByte(length);
        const result = this.buffer.slice(this.offset, this.offset + length);
        this.offset += length;

        return result;
    }

    // gRPC has some additional header - remove it
    trySkipGrpcHeader() {
        const backupOffset = this.offset;

        if (this.buffer[this.offset] === 0 && this.leftBytes() >= 5) {
            this.offset++;
            const length = this.buffer.readInt32BE(this.offset);
            this.offset += 4;

            if (length > this.leftBytes()) {
                // Something is wrong, revert
                this.offset = backupOffset;
            }
        }
    }

    leftBytes() {
        return this.buffer.length - this.offset;
    }

    checkByte(length) {
        const bytesAvailable = this.leftBytes();
        if (length > bytesAvailable) {
            throw new Error('Not enough bytes left. Requested: ' + length + ' left: ' + bytesAvailable);
        }
    }

    checkpoint() {
        this.savedOffset = this.offset;
    }

    resetToCheckpoint() {
        this.offset = this.savedOffset;
    }
}

export const TYPES = {
    VARINT: 0,
    FIXED64: 1,
    STRING: 2,
    FIXED32: 5
};

export function typeToString(type) {
    switch (type) {
        case TYPES.VARINT:
            return 'varint';
        case TYPES.STRING:
            return 'string';
        case TYPES.FIXED32:
            return 'fixed32';
        case TYPES.FIXED64:
            return 'fixed64';
        default:
            return 'unknown';
    }
}
