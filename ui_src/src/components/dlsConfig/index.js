// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server
import './style.scss';
import React, { useState } from 'react';
import TitleComponent from '../titleComponent';
import Switcher from '../switcher';

const DlsConfig = () => {
    const [dlsTypes, setDlsTypes] = useState({
        poison: true,
        schemaverse: true
    });

    const handlePoisonChange = () => {
        setDlsTypes({ ...dlsTypes, poison: !dlsTypes.poison });
    };
    const handleSchemaChange = () => {
        setDlsTypes({ ...dlsTypes, schemaverse: !dlsTypes.schemaverse });
    };

    return (
        <div className="dls-config-container">
            <div className="toggle-dls-config">
                <TitleComponent headerTitle="Poison" typeTitle="sub-header" headerDescription="Contrary to popular belief, Lorem Ipsum is not " />
                <Switcher onChange={handlePoisonChange} checked={dlsTypes.poison} />
            </div>
            <div className="toggle-dls-config">
                <TitleComponent headerTitle="Schemaverse" typeTitle="sub-header" headerDescription="Contrary to popular belief, Lorem Ipsum is not " />
                <Switcher onChange={handleSchemaChange} checked={dlsTypes.schemaverse} />
            </div>
        </div>
    );
};
export default DlsConfig;
