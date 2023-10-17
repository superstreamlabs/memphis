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

import { ReactComponent as StatusIcon } from '../../../../../../assets/images/statusIcon.svg';
import Tag from '../../../../../../components/tag';
import { ColorPalette } from '../../../../../../const/globalConst';

const TestResult = ({ name }) => {
    return (
        <div className="result-wrapper">
            <div className="header">
                <p className="title">EXECUTION RESULT</p>

                <div className="rightSide">
                    <StatusIcon />

                    <span className="status">Status:</span>
                    <Tag tag={{ name: 'Successful', color: ColorPalette[9] }} editable={false} rounded={false} />
                </div>
            </div>
            <div className="result">
                <p>
                    <strong>Test Event Name</strong>{' '}
                </p>
                <p>{name}</p>
                <br />
                <p>
                    <strong>Responses</strong>
                </p>
                <p>
                    {`"errorTypee: "SyntaxError"
"errorMessage": "Unexpected token u in JSON at position O" ,
trace"
"SyntaxError: Unexpected token u in at position ø" ,
at JSON. parse ,
at Runtime *handler
at Runtime *handleOnceNonStreaming`}
                </p>
            </div>
        </div>
    );
};

export default TestResult;
