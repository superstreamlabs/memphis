import { bufferToPrettyHex, parseInput } from './decoderFiles/hexUtils';
import { TYPES, typeToString } from './decoderFiles/protobufDecoder';
import { decodeVarintParts, decodeFixed64, decodeFixed32 } from './decoderFiles/protobufPartDecoder';
import { BufferReader } from './decoderFiles/protobufDecoder';
import { v4 } from 'uuid';

let jsonarray = [];

function checkIsBuffer(node) {
    let n;
    if (node.value._isBuffer) {
        n = {
            index: node.index,
            type: node.type,
            children: node.children
        };
    } else {
        n = {
            index: node.index,
            type: node.type,
            value: node.value,
            children: node.children
        };
    }
    return n;
}

function list_to_tree(list) {
    var map = {},
        node,
        roots = [],
        i;

    for (i = 0; i < list.length; i += 1) {
        map[list[i].uuid] = i; // initialize the map
        list[i].children = []; // initialize the children
    }

    for (i = 0; i < list.length; i += 1) {
        node = list[i];
        if (node.parentId !== null) {
            list[map[node.parentId]].children.push(checkIsBuffer(node));
        } else {
            roots.push(checkIsBuffer(node));
        }
    }
    return roots;
}

function decodeProto(buffer, id = null) {
    const reader = new BufferReader(buffer);
    const parts = [];
    reader.trySkipGrpcHeader();

    try {
        while (reader.leftBytes() > 0) {
            reader.checkpoint();

            const indexType = parseInt(reader.readVarInt().toString());
            const type = indexType & 0b111;
            const index = indexType >> 3;
            const uuid = v4();
            const parentId = id ? id : null;
            const children = null;
            let value;
            if (type === TYPES.VARINT) {
                value = reader.readVarInt().toString();
            } else if (type === TYPES.STRING) {
                const length = parseInt(reader.readVarInt().toString());
                value = reader.readBuffer(length);
            } else if (type === TYPES.FIXED32) {
                value = reader.readBuffer(4);
            } else if (type === TYPES.FIXED64) {
                value = reader.readBuffer(8);
            } else {
                throw new Error('Unknown type: ' + type);
            }

            parts.push({
                uuid,
                parentId,
                index,
                type,
                value
            });
            jsonarray.push({
                uuid,
                parentId,
                index,
                children,
                type: typeToString(type),
                value
            });
        }
    } catch (err) {
        reader.resetToCheckpoint();
    }

    return {
        parts,
        leftOver: reader.readBuffer(reader.leftBytes())
    };
}

function ProtobufStringPart(value, id) {
    const decoded = decodeProto(value, id);
    if (value.length > 0 && decoded.leftOver.length === 0) {
        ProtobufDisplay(decoded);
    } else {
        let index = jsonarray.findIndex((item) => item.uuid === id);
        jsonarray[index].value = value.toString();
        value.toString();
    }
}

function getProtobufPart(part) {
    let decoded;
    switch (part.type) {
        case TYPES.VARINT:
            decoded = decodeVarintParts(part.value);
            break;
        case TYPES.STRING:
            ProtobufStringPart(part.value, part.uuid);
            break;
        case TYPES.FIXED64:
            decoded = decodeFixed64(part.value);
            break;
        case TYPES.FIXED32:
            decoded = decodeFixed32(part.value);
            break;
        default:
            return 'Unknown type';
    }
}

function ProtobufPart(part) {
    const stringType = typeToString(part.type);
    getProtobufPart(part);
}

function ProtobufDisplay(value) {
    value.parts.map((part, i) => {
        ProtobufPart(part);
    });
    // const leftOver = value.leftOver.length ? <p>Left over bytes: {bufferToPrettyHex(value.leftOver)}</p> : null;
}

export const decodeMessage = (message) => {
    jsonarray = [];
    let msg = message.replaceAll(`"`, '');
    const buffer = parseInput(msg);
    let value = decodeProto(buffer);
    ProtobufDisplay(value);
    return list_to_tree(jsonarray);
};
