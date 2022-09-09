// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import React, { useState } from 'react';

import Switcher from '../../../components/switcher';

const Alerts = () => {
    const [errorsAlert, setErrorsAlert] = useState(false);
    const [schemaAlert, setSchemaAlert] = useState(false);
    return (
        <div className="alerts-integrations-container">
            <h3 className="title">We will keep an eye on your data streams and alert you if anything went wrong according to the following triggers:</h3>
            <div>
                <div className="alert-integration-type">
                    <label className="alert-label-bold">Errors</label>
                    <Switcher onChange={() => setErrorsAlert(!errorsAlert)} checked={errorsAlert} checkedChildren="on" unCheckedChildren="off" />
                </div>
                <div className="alert-integration-type">
                    <label className="alert-label-bold">Schema has changed</label>
                    <Switcher onChange={() => setSchemaAlert(!schemaAlert)} checked={schemaAlert} checkedChildren="on" unCheckedChildren="off" />
                </div>
            </div>
        </div>
    );
};

export default Alerts;
