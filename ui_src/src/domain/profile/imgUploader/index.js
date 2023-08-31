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

import React, { useContext, useState } from 'react';
import { Context } from '../../../hooks/store';
import { LOCAL_STORAGE_COMPANY_LOGO } from '../../../const/localStorageConsts';
import { Upload } from 'antd';
import Button from '../../../components/button';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { httpRequest } from '../../../services/http';
import Logo from '../../../assets/images/logo.svg';
import { showMessages } from '../../../services/genericServices';

function ImgUploader() {
    const [state, dispatch] = useContext(Context);
    const [fileList, setFileList] = useState(
        localStorage.getItem(LOCAL_STORAGE_COMPANY_LOGO)
            ? [
                  {
                      uid: '1',
                      name: 'company_logo',
                      status: 'done',
                      url: localStorage.getItem(LOCAL_STORAGE_COMPANY_LOGO)
                  }
              ]
            : []
    );

    const props = {
        beforeUpload: (file) => {
            const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
            if (!isJpgOrPng) {
                showMessages('error', 'You can only upload JPG/PNG file!');
            }
            setFileList([file]);
            return isJpgOrPng;
        },
        customRequest: (file) => uploadLogo(file),
        fileList
    };

    const uploadLogo = async ({ file, onSuccess, onError }) => {
        let dataImg = new FormData();
        dataImg.append('file', file);
        try {
            const data = await httpRequest('PUT', ApiEndpoints.EDIT_COMPANY_LOGO, dataImg);
            localStorage.setItem(LOCAL_STORAGE_COMPANY_LOGO, data.image);
            dispatch({ type: 'SET_COMPANY_LOGO', payload: data.image });
            onSuccess('ok');
        } catch (err) {
            onError('error');
        }
    };

    const deleteLogo = async ({ onSuccess, onError }) => {
        try {
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_COMPANY_LOGO);
            localStorage.setItem(LOCAL_STORAGE_COMPANY_LOGO, null);
            dispatch({ type: 'SET_COMPANY_LOGO', payload: null });
            setFileList([]);
            onSuccess('ok');
        } catch (err) {
            onError('error');
        }
    };

    return (
        <div className="company-logo-section">
            <p className="title">Company Logo</p>
            <div className="company-logo">
                <img className="logoimg" src={state?.companyLogo || Logo} alt="companyLogo" />
                <div className="company-logo-right">
                    <div className="update-remove-logo">
                        <Upload {...props} name="company-logo" maxCount={1} showUploadList={false} fileList={fileList}>
                            <Button
                                className="modal-btn"
                                width="160px"
                                height="36px"
                                placeholder="Upload New"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="14px"
                                fontWeight="600"
                                aria-haspopup="true"
                            />
                        </Upload>
                        <Button
                            className="modal-btn"
                            width="200px"
                            height="36px"
                            placeholder="Remove Current Logo"
                            colorType="red"
                            radiusType="circle"
                            backgroundColorType="none"
                            border="gray"
                            boxShadowsType="gray"
                            fontSize="14px"
                            fontWeight="600"
                            aria-haspopup="true"
                            onClick={() => deleteLogo(fileList[0])}
                            disabled={!state?.companyLogo}
                        />
                    </div>
                    <label className="company-logo-description">Logo must be 200x200 pixel and size is less than 5mb</label>
                </div>
            </div>
        </div>
    );
}

export default ImgUploader;
