// Credit for The NATS.IO Authors
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

import { CloseRounded } from '@material-ui/icons';
import { Drawer, Button, Space } from 'antd';
import React, { useState } from 'react';

import createdDateIcon from '../../../../assets/images/createdDateIcon.svg';
import notUsedIcond from '../../../../assets/images/notUsedIcon.svg';
import { parsingDate } from '../../../../services/valueConvertor';
import CheckboxComponent from '../../../../components/checkBox';
import usedIcond from '../../../../assets/images/usedIcon.svg';
import TagsList from '../../../../components/tagList';
import SchemaDetails from '../schemaDetails';
import TooltipComponent from '../../../../components/tooltip/tooltip';
import OverflowTip from '../../../../components/tooltip/overflowtip';

function SchemaBox({ schema, handleCheckedClick, isCheck }) {
    const [open, setOpen] = useState(false);

    const handleDrawer = (flag) => {
        setOpen(flag);
    };

    return (
        <>
            <div>
                <CheckboxComponent checked={isCheck} id={schema.name} onChange={handleCheckedClick} name={schema.name} />
                <div key={schema.name} onClick={() => handleDrawer(true)} className="schema-box-wrapper">
                    <header is="x3d">
                        <div className="header-wrapper">
                            <div className="schema-name">
                                <OverflowTip text={schema.name} maxWidth={'100px'}>
                                    <span>{schema.name}</span>
                                </OverflowTip>
                            </div>
                            <div className="is-used">
                                <img src={schema.used ? usedIcond : notUsedIcond} alt="usedIcond" />
                                {schema.used && <p className="used">Used</p>}
                                {!schema.used && <p className="not-used"> Not used</p>}
                            </div>
                        </div>
                    </header>
                    <type is="x3d">
                        <div className="field-wrapper">
                            <p>Type : </p>
                            {schema.type === 'json' ? <p className="schema-json-name">JSON schema</p> : <span> {schema.type}</span>}
                        </div>
                        <div className="field-wrapper">
                            <p>Created by : </p>
                            <OverflowTip text={schema.created_by_user} maxWidth={'100px'}>
                                <span>{schema.created_by_user}</span>
                            </OverflowTip>
                        </div>
                    </type>
                    <tags is="x3d">
                        <TagsList tagsToShow={3} tags={schema?.tags} />
                    </tags>
                    <date is="x3d">
                        <img src={createdDateIcon} alt="createdDateIcon" />
                        <p>{parsingDate(schema.creation_date)}</p>
                    </date>
                </div>
            </div>
            <Drawer
                title={schema?.name}
                placement="right"
                size={'large'}
                onClose={() => handleDrawer(false)}
                open={open}
                maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}
                closeIcon={<CloseRounded style={{ color: '#D1D1D1' }} />}
            >
                <SchemaDetails schemaName={schema?.name} closeDrawer={() => handleDrawer(false)} />
            </Drawer>
        </>
    );
}

export default SchemaBox;
