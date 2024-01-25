import { parseInput } from './decoderFiles/hexUtils';
import { TYPES, typeToString } from './decoderFiles/protobufDecoder';
import { BufferReader } from './decoderFiles/protobufDecoder';
import { v4 } from 'uuid';

let jsonarray = [];

const pushToArray = (part) => {
    jsonarray.push({
        uuid: part.uuid,
        parentId: part.parentId,
        index: part.index,
        type: typeToString(part.type),
        value: part.value
    });
};

function checkIsBuffer(node) {
    let n;
    if (node.value._isBuffer) {
        n = {
            field_number: node.index,
            type: node.type,
            children: node.children
        };
    } else {
        n = {
            field_number: node.index,
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
        list[i].children = []; // initialize the childrens
    }

    for (i = 0; i < list.length; i += 1) {
        node = list[i];
        if (node.parentId !== null) {
            list[map[node.parentId]].children.push(checkIsBuffer(node));
        } else {
            roots.push(checkIsBuffer(node));
        }
    }
    for (i = 0; i < roots.length; i += 1) {
        if (roots[i].children.length === 0 && !roots[i].value) {
            roots.splice(i, 1);
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
        }
    } catch (err) {
        reader.resetToCheckpoint();
    }
    return {
        parts,
        leftOver: reader.readBuffer(reader.leftBytes())
    };
}

function ProtobufStringPart(part) {
    if (part.parentId === null) {
        pushToArray(part);
    }
    const decoded = decodeProto(part.value, part.uuid);
    if (part.value.length > 0 && decoded.leftOver.length === 0) {
        pushToArray(part);
        ProtobufDisplay(decoded);
    } else {
        if (part.parentId) {
            pushToArray(part);
        }
        let index = jsonarray.findIndex((item) => item.uuid === part.uuid);
        if (index !== -1) {
            jsonarray[index].value = part.value.toString();
        }
    }
}

function ProtobufDisplay(value) {
    value.parts.map((part) => {
        if (part.type === TYPES.STRING) {
            ProtobufStringPart(part);
        } else {
            pushToArray(part);
        }
    });
}

export const decodeMessage = (message) => {
    jsonarray = [];
    let msg = message.replaceAll(`"`, '');
    const buffer = parseInput(msg);
    let value = decodeProto(buffer);
    ProtobufDisplay(value);
    return list_to_tree(jsonarray);
};
