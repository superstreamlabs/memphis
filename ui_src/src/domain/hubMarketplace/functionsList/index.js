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

import { SearchOutlined } from '@ant-design/icons';
import { Pagination } from 'antd';
import React from 'react';

import SearchInput from '../../../components/searchInput';
import SelectComponent from '../../../components/select';
import { ReactComponent as InstalledIcon } from '../../../assets/images/installed.svg';
import VerifiedBedge from '../verifiedBedge';

const funcList = [
    {
        id: 1,
        funcName: 'Sentiment analysis very long test and more',
        funcDesc:
            'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi feugiat commodo mauris, eget imperdiet nisl interdum vel.nLorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi feugiat commodo mauris, eget imperdiet nisl interdum vel.',
        inputDataType: 'JSON',
        outputDataType: 'JSON',
        isVerified: true,
        funcImg: 'https://gw.alipayobjects.com/zos/antfincdn/LlvErxo8H9/photo-1503185912284-5271ff81b9a8.webp'
    },
    {
        id: 2,
        funcName: 'Sentiment analysis',
        funcDesc:
            'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Proin facilisis arcu nec arcu ultrices, ipsum dolor sit amet, consectetur adipiscing elit.Proin facilisis arcu nec arcu ultrices, ipsum dolor sit amet, consectetur adipiscing elit.Proin facilisis arcu nec arcu ultrices, ipsum dolor sit amet, consectetur adipiscing elit.Proin facilisis arcu nec arcu ultrices, ipsum dolor sit amet, consectetur adipiscing elit.Proin facilisis arcu nec arcu ultrices, ipsum dolor sit amet, consectetur adipiscing elit. Proin facilisis arcu nec arcu ultrices, sit amet lacinia libero hendrerit. Maecenas sollicitudin eros nulla, eu porttitor enim convallis sed. Vivamus finibus interdum bibendum.\n Nullam a sodales eros, quis facilisis arcu. Curabitur rutrum auctor volutpat. Donec purus orci, mattis auctor gravida sit amet, tincidunt in nunc. Praesent et magna varius, egestas felis id, scelerisque mi. Nam lectus leo, tempor ut tristique sed, condimentum ac leo. Integer venenatis auctor elit, ac vestibulum magna sollicitudin mattis. Nulla facilisi. Nunc imperdiet nisi id dolor tincidunt auctor. Nunc maximus sapien sit amet quam placerat, non hendrerit odio dapibus. Integer semper congue quam at facilisis. Nam consectetur consectetur velit. Donec scelerisque nec ante in euismod. Morbi ut elit accumsan, congue enim consectetur, ultrices mauris.',
        inputDataType: 'Text',
        outputDataType: 'JSON',
        isVerified: true,
        funcImg: null
    },
    {
        id: 4,
        funcName: 'Sentiment analysis very long test and more',
        funcDesc: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi feugiat commodo mauris, eget imperdiet nisl interdum vel.',
        inputDataType: 'Text',
        outputDataType: 'JSON',
        isVerified: true,
        funcImg: null
    },
    {
        id: 3,
        funcName: 'Sentiment analysis',
        funcDesc: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi feugiat commodo mauris, eget imperdiet nisl interdum vel.',
        inputDataType: 'XML',
        outputDataType: 'JSON',
        isVerified: false,
        funcImg: 'https://gw.alipayobjects.com/zos/antfincdn/x43I27A55%26/photo-1438109491414-7198515b166b.webp'
    }
];

function FunctionsList(props) {
    const [activeFunction, setActiveFunction] = React.useState(null);

    const chooseFunction = (func) => {
        setActiveFunction(func.id);
        props.chooseFunction(func);
    };

    return (
        <div className="function-list">
            <div className="function-list-header">
                <p>Functions</p>
                <div>
                    <SelectComponent
                        value="val0"
                        colorType="navy"
                        backgroundColorType="none"
                        borderColorType="gray"
                        radiusType="circle"
                        width="10vw"
                        height="27px"
                        options={['op1', 'op2']}
                        boxShadowsType="gray"
                        popupClassName="select-options"
                    />
                    <SearchInput
                        placeholder="Search here"
                        colorType="navy"
                        backgroundColorType="none"
                        width="10vw"
                        height="27px"
                        borderRadiusType="circle"
                        borderColorType="gray"
                        boxShadowsType="gray"
                        iconComponent={<SearchOutlined />}
                        //   onChange={handleSearch}
                        //   value={searchInput}
                    />
                </div>
            </div>
            <div className="function-list-body">
                {funcList &&
                    funcList.map((func) => (
                        <div
                            className="function-item-container"
                            key={func.id}
                            style={{ border: func.id === activeFunction && '1px solid #5A4FE5' }}
                            onClick={() => chooseFunction(func)}
                        >
                            <div className="function-details">
                                {func.funcImg ? (
                                    <img src={func.funcImg} alt="function" width="50" height="50" className="img-placeholder" />
                                ) : (
                                    <div className="img-placeholder" />
                                )}
                                <div>
                                    <p className="function-name">{func.funcName}</p>
                                    <p className="data-type">Data type: {func.inputDataType}</p>
                                    {func.isVerified && <VerifiedBedge />}
                                </div>
                            </div>
                            <div className="function-description">{func.funcDesc}</div>
                            <InstalledIcon alt="Downloaded" width={20} height={20} className="installed" />
                        </div>
                    ))}
            </div>
            <div className="function-list-footer">
                <Pagination
                    // total={funcList.length}
                    total={100}
                    // showSizeChanger
                    showQuickJumper
                    // showTotal={total => `Total ${total} items`}
                />
            </div>
        </div>
    );
}
export default FunctionsList;
