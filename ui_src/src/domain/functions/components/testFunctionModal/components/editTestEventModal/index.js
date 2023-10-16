// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import { useState } from 'react';
import Button from '../../../../../../components/button';
import { FaPlay } from 'react-icons/fa';
import Editor from '@monaco-editor/react';
import TestResult from '../testResult';

const DEFAULT_TEXT = `"type": "record",
"namespace": "com.example",
"name": "test-schema",
"fields": [
       { "name": "Master message", "type": "string", "default": "NONE" },
       { "name": "age", "type": "int", "default": "-1" },
       { "name": "phone", "type": "string", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
]`;
const EditTestEventModal = ({ event, handleEdit }) => {
    const [isTested, setIsTested] = useState(false);

    return (
        <div className="editTestEventModal-container">
            <div className="actions-wrapper">
                <p className="title">Event data</p>
                <div className="actions">
                    <Button
                        placeholder={'Edit'}
                        backgroundColorType={'purple'}
                        border={'none'}
                        colorType={'white'}
                        fontSize={'12px'}
                        fontFamily={'InterSemiBold'}
                        radiusType={'circle'}
                        height={'34px'}
                        width={'70px'}
                        onClick={() => handleEdit()}
                    />
                    <Button
                        placeholder={
                            <div className="button-content">
                                <FaPlay />
                                <span>Test</span>
                            </div>
                        }
                        backgroundColorType={'orange'}
                        border={'none'}
                        colorType={'black'}
                        fontSize={'12px'}
                        fontFamily={'InterSemiBold'}
                        radiusType={'circle'}
                        height={'34px'}
                        width={'99px'}
                        onClick={() => setIsTested(true)}
                    />
                </div>
            </div>
            <div className="text-area-wrapper">
                <div className={`text-wrapper ${isTested ? 'width-50' : undefined}  `}>
                    <Editor
                        options={{
                            minimap: { enabled: false },
                            scrollbar: { verticalScrollbarSize: 3 },
                            scrollBeyondLastLine: false,
                            roundedSelection: false,
                            formatOnPaste: true,
                            formatOnType: true,
                            fontSize: '14px',
                            fontFamily: 'Inter',
                            lineNumbers: 'off',
                            glyphMargin: false,
                            folding: false,
                            lineDecorationsWidth: 0,
                            lineNumbersMinChars: 0,
                            readOnly: true,
                            automaticLayout: true
                        }}
                        language={'json'}
                        height="100%"
                        defaultValue={DEFAULT_TEXT}
                        value={event.content}
                        key={isTested ? 'tested' : 'not-tested'}
                    />
                </div>
                {isTested && <TestResult name={event.name} />}
            </div>
        </div>
    );
};

export default EditTestEventModal;
